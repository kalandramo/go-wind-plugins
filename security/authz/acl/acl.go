// Package acl implements a lightweight [engine.Engine] based on simple
// Access Control Lists.
//
// Rules are evaluated in order. Each rule specifies a subject, action,
// resource (with wildcard support), and an effect (allow/deny).
// Wildcards: "*" matches any value; "resource:*" matches all actions on
// the resource; "doc:*" matches doc:read, doc:write, etc.
//
// By default, access is denied if no rule matches (defaultDeny=true),
// and deny rules override allow rules (denyOverrides=true).
package acl

import (
	"context"
	"strings"
	"sync"

	engine "github.com/tx7do/go-wind-plugins/security/authz"
)

// State implements the ACL authz engine.
type State struct {
	mu      sync.RWMutex
	options *Options
}

func init() {
	_ = engine.Register(engine.Acl, func(ctx context.Context, options ...any) (engine.Engine, error) {
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

// NewEngine creates an ACL engine from the given options.
func NewEngine(_ context.Context, opts ...OptFunc) (*State, error) {
	o := &Options{
		wildcard:      "*",
		defaultDeny:   true,
		denyOverrides: true,
	}
	for _, opt := range opts {
		opt(o)
	}
	return &State{options: o}, nil
}

func (s *State) Name() string {
	return string(engine.Acl)
}

func (s *State) IsAuthorized(_ context.Context, subject engine.Subject, action engine.Action, resource engine.Resource, _ engine.Project) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.check(string(subject), string(action), string(resource)), nil
}

func (s *State) ProjectsAuthorized(_ context.Context, subjects engine.Subjects, action engine.Action, resource engine.Resource, projects engine.Projects) (engine.Projects, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(engine.Projects, 0, len(projects))
	for _, project := range projects {
		for _, sub := range subjects {
			if s.check(string(sub), string(action), string(resource)) {
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
			if s.check(string(sub), string(p.Action), string(p.Resource)) {
				result = append(result, p)
				break
			}
		}
	}
	return result, nil
}

func (s *State) FilterAuthorizedProjects(_ context.Context, subjects engine.Subjects) (engine.Projects, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// With ACL there are no "projects" concept; return empty.
	return engine.Projects{}, nil
}

func (s *State) SetPolicies(_ context.Context, policyMap engine.PolicyMap, _ engine.RoleMap) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Allow policies to be provided as a []Rule slice.
	if rules, ok := policyMap["rules"].([]Rule); ok {
		s.options.rules = rules
	}
	return nil
}

// check evaluates whether the given subject/action/resource is allowed.
func (s *State) check(subject, action, resource string) bool {
	allowed := false
	denied := false

	for _, rule := range s.options.rules {
		if !matchValue(rule.Subject, subject, s.options.wildcard) {
			continue
		}
		if !matchValue(rule.Action, action, s.options.wildcard) {
			continue
		}
		if !matchValue(rule.Resource, resource, s.options.wildcard) {
			continue
		}

		if rule.Effect == "deny" {
			denied = true
		} else {
			// Default to allow.
			allowed = true
		}
	}

	if s.options.denyOverrides && denied {
		return false
	}
	if allowed {
		return true
	}
	return !s.options.defaultDeny
}

// matchValue checks if a pattern matches a value, supporting wildcard '*' for
// prefix and suffix matching.
//
//	"*" matches everything
//	"doc:*" matches "doc:read", "doc:write", etc.
//	"doc:read" matches only "doc:read"
func matchValue(pattern, value, wildcard string) bool {
	if pattern == wildcard {
		return true
	}
	if pattern == value {
		return true
	}

	// Prefix wildcard: "*:read" matches "doc:read", "img:read"
	if strings.HasPrefix(pattern, wildcard) {
		return strings.HasSuffix(value, pattern[1:])
	}

	// Suffix wildcard: "doc:*" matches "doc:read", "doc:write"
	if strings.HasSuffix(pattern, wildcard) {
		return strings.HasPrefix(value, pattern[:len(pattern)-1])
	}

	return false
}
