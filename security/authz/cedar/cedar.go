// Package cedar implements an [engine.Engine] that delegates authorization
// decisions to AWS Verified Permissions, which evaluates Cedar policies.
//
// Cedar is Amazon's open-source policy language designed for authorization.
// This engine communicates with the AWS Verified Permissions service via its
// HTTP API (IsAuthorized endpoint).
//
// For local/testing use, provide WithAuthorizer to bypass the remote call.
//
// Usage:
//
//	e, _ := cedar.NewEngine(ctx,
//	    cedar.WithPolicyStoreID("store-123"),
//	    cedar.WithRegion("us-east-1"),
//	)
//	ok, _ := e.IsAuthorized(ctx, "alice", "read", "doc:1", "")
package cedar

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	engine "github.com/tx7do/go-wind-plugins/security/authz"
)

// State implements the Cedar authz engine.
type State struct {
	mu      sync.RWMutex
	options *Options
}

func init() {
	_ = engine.Register(engine.Cedar, func(ctx context.Context, options ...any) (engine.Engine, error) {
		var opts []OptFunc
		for _, o := range options {
			if opt, ok := o.(OptFunc); ok {
				opts = append(opts, opt)
			}
		}
		return NewEngine(ctx, opts...)
	})
}

var _ engine.Engine = (*State)(nil)

// NewEngine creates a Cedar engine from the given options.
func NewEngine(_ context.Context, opts ...OptFunc) (*State, error) {
	o := &Options{}
	for _, opt := range opts {
		opt(o)
	}
	return &State{options: o}, nil
}

func (s *State) Name() string {
	return string(engine.Cedar)
}

func (s *State) IsAuthorized(ctx context.Context, subject engine.Subject, action engine.Action, resource engine.Resource, _ engine.Project) (bool, error) {
	return s.check(ctx, string(subject), string(action), string(resource))
}

func (s *State) ProjectsAuthorized(ctx context.Context, subjects engine.Subjects, action engine.Action, resource engine.Resource, projects engine.Projects) (engine.Projects, error) {
	result := make(engine.Projects, 0, len(projects))
	for _, project := range projects {
		for _, sub := range subjects {
			ok, err := s.check(ctx, string(sub), string(action), string(resource))
			if err != nil {
				return nil, err
			}
			if ok {
				result = append(result, project)
				break
			}
		}
	}
	return result, nil
}

func (s *State) FilterAuthorizedPairs(ctx context.Context, subjects engine.Subjects, pairs engine.Pairs) (engine.Pairs, error) {
	result := make(engine.Pairs, 0, len(pairs))
	for _, p := range pairs {
		for _, sub := range subjects {
			ok, err := s.check(ctx, string(sub), string(p.Action), string(p.Resource))
			if err != nil {
				return nil, err
			}
			if ok {
				result = append(result, p)
				break
			}
		}
	}
	return result, nil
}

func (s *State) FilterAuthorizedProjects(_ context.Context, _ engine.Subjects) (engine.Projects, error) {
	return engine.Projects{}, nil
}

func (s *State) SetPolicies(_ context.Context, _ engine.PolicyMap, _ engine.RoleMap) error {
	// Policies are managed in the AVP service, not locally.
	return nil
}

// check evaluates a single authorization request.
func (s *State) check(ctx context.Context, principal, action, resource string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Local authorizer (for testing or custom Cedar evaluators).
	if s.options.authorizer != nil {
		return s.options.authorizer(principal, action, resource)
	}

	// Remote call to AWS Verified Permissions.
	if s.options.policyStoreID == "" {
		return false, fmt.Errorf("policy store ID is required for remote evaluation")
	}

	return s.callAVP(ctx, principal, action, resource)
}

// callAVP sends an IsAuthorized request to AWS Verified Permissions.
// Note: In production, this requires AWS SigV4 signing.
func (s *State) callAVP(ctx context.Context, principal, action, resource string) (bool, error) {
	reqBody := map[string]interface{}{
		"policyStoreId": s.options.policyStoreID,
		"principal": map[string]string{
			"entityType": "User",
			"entityId":   principal,
		},
		"action": map[string]string{
			"entityType": "Action",
			"entityId":   action,
		},
		"resource": map[string]string{
			"entityType": "Resource",
			"entityId":   resource,
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return false, fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := s.options.getEndpoint() + "/isauthorized"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(jsonBody))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/x-amz-json-1.0")
	req.Header.Set("X-Amz-Target", "AVPControlService.IsAuthorized")

	// Note: AWS SigV4 signing should be applied here in production.
	// Users can inject a custom HTTP client that handles signing.

	resp, err := s.options.getHTTPClient().Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("AVP returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Decision string `json:"decision"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("failed to decode AVP response: %w", err)
	}

	return result.Decision == "ALLOW", nil
}
