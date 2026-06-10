// Package cerbos implements an [engine.Engine] that delegates authorization
// decisions to a Cerbos Policy Decision Point (PDP) via its HTTP API.
//
// Cerbos is an open-source, stateless authorization service that evaluates
// policies written in YAML/JSON. It is purpose-built for authorization,
// offering better performance and simplicity compared to general-purpose
// policy engines.
//
// Usage:
//
//	e, _ := cerbos.NewEngine(ctx, cerbos.WithEndpoint("http://localhost:3592"))
//	ok, _ := e.IsAuthorized(ctx, "alice", "read", "doc:1", "")
//
// For testing, use WithAuthorizer to bypass the remote call.
package cerbos

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

// State implements the Cerbos authz engine.
type State struct {
	mu      sync.RWMutex
	options *Options
}

func init() {
	_ = engine.Register(engine.Cerbos, func(ctx context.Context, options ...any) (engine.Engine, error) {
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

// NewEngine creates a Cerbos engine from the given options.
func NewEngine(_ context.Context, opts ...OptFunc) (*State, error) {
	o := &Options{}
	for _, opt := range opts {
		opt(o)
	}
	return &State{options: o}, nil
}

func (s *State) Name() string {
	return string(engine.Cerbos)
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
	// Policies are managed by the Cerbos PDP, not locally.
	return nil
}

// check evaluates a single authorization request.
func (s *State) check(ctx context.Context, principal, action, resource string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Local authorizer (for testing).
	if s.options.authorizer != nil {
		return s.options.authorizer(principal, action, resource)
	}

	return s.callCerbos(ctx, principal, action, resource)
}

// callCerbos sends a CheckResources request to the Cerbos PDP.
func (s *State) callCerbos(ctx context.Context, principal, action, resource string) (bool, error) {
	// Build the Cerbos CheckResources API request.
	roles := []string{"user"}
	if r, ok := s.options.principalRoles[principal]; ok && len(r) > 0 {
		roles = r
	}

	// Split resource "doc:1" into kind:id
	resKind, resID := splitResource(resource)

	reqBody := map[string]interface{}{
		"requestId": "authz-check",
		"principal": map[string]interface{}{
			"id":    principal,
			"roles": roles,
		},
		"resources": []map[string]interface{}{
			{
				"resource": map[string]interface{}{
					"kind": resKind,
					"id":   resID,
				},
				"actions": []string{action},
			},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return false, fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := s.options.getEndpoint() + "/api/check/resources"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(jsonBody))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.options.getHTTPClient().Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("cerbos returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Results []struct {
			Actions map[string]bool `json:"actions"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("failed to decode cerbos response: %w", err)
	}

	if len(result.Results) == 0 {
		return false, nil
	}

	return result.Results[0].Actions[action], nil
}

// splitResource splits "doc:1" into ("doc", "1").
// If no colon, the entire string is used as kind with id "default".
func splitResource(resource string) (kind, id string) {
	for i := len(resource) - 1; i >= 0; i-- {
		if resource[i] == ':' {
			return resource[:i], resource[i+1:]
		}
	}
	return resource, "default"
}
