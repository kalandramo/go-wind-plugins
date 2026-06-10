// Package rbac implements a lightweight [engine.Engine] based on
// Role-Based Access Control.
//
// The engine maintains two maps:
//
//	user → roles
//	role → permissions (resource:action pairs)
//
// Wildcard support: "*" in a permission matches any value.
// Role inheritance is supported via assigning parent roles as users
// (e.g. WithUserRole("admin", "editor")).
package rbac

import (
	"context"
	"strings"
	"sync"

	engine "github.com/tx7do/go-wind-plugins/security/authz"
)

// State implements the RBAC authz engine.
type State struct {
	mu      sync.RWMutex
	options *Options
}

func init() {
	_ = engine.Register(engine.Rbac, func(ctx context.Context, options ...any) (engine.Engine, error) {
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

// NewEngine creates an RBAC engine from the given options.
func NewEngine(_ context.Context, opts ...OptFunc) (*State, error) {
	o := &Options{
		wildcard:        "*",
		rolePermissions: make(map[string][]Permission),
		userRoles:       make(map[string][]string),
	}
	for _, opt := range opts {
		opt(o)
	}
	return &State{options: o}, nil
}

func (s *State) Name() string {
	return string(engine.Rbac)
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

func (s *State) FilterAuthorizedProjects(_ context.Context, _ engine.Subjects) (engine.Projects, error) {
	return engine.Projects{}, nil
}

func (s *State) SetPolicies(_ context.Context, policyMap engine.PolicyMap, roleMap engine.RoleMap) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Load role → permissions from policyMap.
	if rp, ok := policyMap["rolePermissions"].(map[string][]Permission); ok {
		s.options.rolePermissions = rp
	}

	// Load user → roles from roleMap.
	if ur, ok := roleMap["userRoles"].(map[string][]string); ok {
		s.options.userRoles = ur
	}

	return nil
}

// check evaluates whether the subject (user) has permission for the
// action/resource via any of its roles.
func (s *State) check(subject, action, resource string) bool {
	// Resolve all roles for the subject (with inheritance).
	roles := s.resolveRoles(subject, make(map[string]bool))

	for _, role := range roles {
		perms := s.options.rolePermissions[role]
		for _, perm := range perms {
			if matchPerm(perm, action, resource, s.options.wildcard) {
				return true
			}
		}
	}
	return false
}

// resolveRoles recursively resolves roles including inheritance.
// A role can be assigned to another "role" to create a hierarchy.
func (s *State) resolveRoles(subject string, visited map[string]bool) []string {
	if visited[subject] {
		return nil
	}
	visited[subject] = true

	roles := s.options.userRoles[subject]
	result := make([]string, 0, len(roles))

	for _, role := range roles {
		result = append(result, role)
		// Check if this role inherits from other roles.
		result = append(result, s.resolveRoles(role, visited)...)
	}

	return result
}

// matchPerm checks if a permission matches the given action/resource.
func matchPerm(perm Permission, action, resource, wildcard string) bool {
	return matchVal(perm.Action, action, wildcard) && matchVal(perm.Resource, resource, wildcard)
}

func matchVal(pattern, value, wildcard string) bool {
	if pattern == wildcard {
		return true
	}
	if pattern == value {
		return true
	}
	if strings.HasSuffix(pattern, wildcard) {
		return strings.HasPrefix(value, pattern[:len(pattern)-1])
	}
	if strings.HasPrefix(pattern, wildcard) {
		return strings.HasSuffix(value, pattern[1:])
	}
	return false
}
