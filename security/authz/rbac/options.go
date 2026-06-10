package rbac

// Permission represents a single resource:action permission.
type Permission struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

// Options holds configuration for the RBAC engine.
type Options struct {
	// rolePermissions maps role → permissions.
	rolePermissions map[string][]Permission

	// userRoles maps user → roles.
	userRoles map[string][]string

	// wildcard is the character used for wildcard matching. Defaults to "*".
	wildcard string
}

type OptFunc func(*Options)

// WithRolePermissions sets the entire role→permissions map.
func WithRolePermissions(rpm map[string][]Permission) OptFunc {
	return func(o *Options) { o.rolePermissions = rpm }
}

// WithRolePermission adds a permission to a role.
func WithRolePermission(role, resource, action string) OptFunc {
	return func(o *Options) {
		if o.rolePermissions == nil {
			o.rolePermissions = make(map[string][]Permission)
		}
		o.rolePermissions[role] = append(o.rolePermissions[role],
			Permission{Resource: resource, Action: action})
	}
}

// WithUserRoles sets the entire user→roles map.
func WithUserRoles(urm map[string][]string) OptFunc {
	return func(o *Options) { o.userRoles = urm }
}

// WithUserRole assigns a role to a user.
func WithUserRole(user, role string) OptFunc {
	return func(o *Options) {
		if o.userRoles == nil {
			o.userRoles = make(map[string][]string)
		}
		o.userRoles[user] = append(o.userRoles[user], role)
	}
}

// WithWildcard sets the wildcard character.
func WithWildcard(w string) OptFunc {
	return func(o *Options) { o.wildcard = w }
}
