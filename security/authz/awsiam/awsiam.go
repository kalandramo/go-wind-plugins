// Package awsiam implements an [engine.Engine] that evaluates authorization
// using AWS IAM-style JSON policies.
//
// IAM policies use a simple deny-by-default model:
//
//	{
//	  "Version": "2012-10-17",
//	  "Statement": [
//	    {
//	      "Effect": "Allow",
//	      "Action": ["s3:GetObject", "s3:ListBucket"],
//	      "Resource": ["arn:aws:s3:::my-bucket/*"]
//	    },
//	    {
//	      "Effect": "Deny",
//	      "Action": ["s3:DeleteObject"],
//	      "Resource": ["arn:aws:s3:::my-bucket/protected/*"]
//	    }
//	  ]
//	}
//
// Wildcard matching follows AWS conventions:
//   - "*" matches everything
//   - "s3:*" matches "s3:GetObject", "s3:PutObject", etc.
//   - "arn:aws:s3:::bucket/*" matches all objects in the bucket
//
// Evaluation rules (same as AWS IAM):
//  1. Default deny (unless WithDefaultAllow is set)
//  2. Explicit Deny always wins
//  3. Allow if any Allow statement matches
package awsiam

import (
	"context"
	"strings"
	"sync"

	engine "github.com/tx7do/go-wind-plugins/security/authz"
)

// State implements the AWS IAM authz engine.
type State struct {
	mu      sync.RWMutex
	options *Options
}

func init() {
	_ = engine.Register(engine.AwsIam, func(ctx context.Context, options ...any) (engine.Engine, error) {
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

// NewEngine creates an AWS IAM engine from the given options.
func NewEngine(_ context.Context, opts ...OptFunc) (*State, error) {
	o := &Options{
		policies:    make(map[string][]Policy),
		defaultDeny: true,
	}
	for _, opt := range opts {
		opt(o)
	}
	return &State{options: o}, nil
}

func (s *State) Name() string {
	return string(engine.AwsIam)
}

func (s *State) IsAuthorized(_ context.Context, subject engine.Subject, action engine.Action, resource engine.Resource, _ engine.Project) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.evaluate(string(subject), string(action), string(resource)), nil
}

func (s *State) ProjectsAuthorized(_ context.Context, subjects engine.Subjects, action engine.Action, resource engine.Resource, projects engine.Projects) (engine.Projects, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(engine.Projects, 0, len(projects))
	for _, project := range projects {
		for _, sub := range subjects {
			if s.evaluate(string(sub), string(action), string(resource)) {
				result = append(result, project)
				break
			}
		}
	}
	return result, nil
}

func (s *State) FilterAuthorizedPairs(_ context.Context, subjects engine.Subjects, pairs engine.Pairs) (engine.Pairs, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(engine.Pairs, 0, len(pairs))
	for _, p := range pairs {
		for _, sub := range subjects {
			if s.evaluate(string(sub), string(p.Action), string(p.Resource)) {
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

func (s *State) SetPolicies(_ context.Context, policyMap engine.PolicyMap, _ engine.RoleMap) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if policies, ok := policyMap["policies"].(map[string][]Policy); ok {
		s.options.policies = policies
	}
	return nil
}

// evaluate checks all statements for the subject and returns the decision.
func (s *State) evaluate(subject, action, resource string) bool {
	policies := s.options.policies[subject]

	allowed := false
	denied := false

	for _, policy := range policies {
		for _, stmt := range policy.Statement {
			if !matchAny(stmt.Actions, action) {
				continue
			}
			if !matchAny(stmt.Resources, resource) {
				continue
			}

			if stmt.Effect == "Deny" {
				denied = true
			} else if stmt.Effect == "Allow" {
				allowed = true
			}
		}
	}

	// Explicit deny always wins.
	if denied {
		return false
	}
	if allowed {
		return true
	}
	return !s.options.defaultDeny
}

// matchAny checks if any pattern matches the value using AWS wildcard rules.
func matchAny(patterns []string, value string) bool {
	for _, p := range patterns {
		if matchPattern(p, value) {
			return true
		}
	}
	return false
}

// matchPattern implements AWS-style wildcard matching:
//   - "*" matches everything
//   - prefix* matches anything starting with prefix
//   - *suffix matches anything ending with suffix
//   - "a*b" matches anything starting with a and ending with b
func matchPattern(pattern, value string) bool {
	if pattern == "*" {
		return true
	}
	if pattern == value {
		return true
	}
	if !strings.Contains(pattern, "*") {
		return pattern == value
	}

	// Split on * and match sequentially.
	parts := strings.Split(pattern, "*")
	if len(parts) == 1 {
		return pattern == value
	}

	pos := 0
	// First part must match prefix.
	if parts[0] != "" {
		if !strings.HasPrefix(value, parts[0]) {
			return false
		}
		pos = len(parts[0])
	}

	// Last part must match suffix.
	lastPart := parts[len(parts)-1]
	if lastPart != "" {
		if !strings.HasSuffix(value, lastPart) {
			return false
		}
	}

	// Middle parts must appear in order.
	for i := 1; i < len(parts)-1; i++ {
		if parts[i] == "" {
			continue
		}
		idx := strings.Index(value[pos:], parts[i])
		if idx < 0 {
			return false
		}
		pos += idx + len(parts[i])
	}

	return true
}
