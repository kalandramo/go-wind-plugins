package acl

// Rule represents a single ACL entry.
type Rule struct {
	Subject string `json:"subject"`
	Action  string `json:"action"`
	// Resource supports wildcard '*' for prefix/suffix matching.
	Resource string `json:"resource"`
	// Effect: "allow" (default) or "deny".
	Effect string `json:"effect,omitempty"`
}

// Options holds configuration for the ACL engine.
type Options struct {
	// rules is the list of ACL rules to evaluate.
	rules []Rule

	// wildcard is the character used for wildcard matching.
	// Defaults to "*".
	wildcard string

	// defaultDeny controls whether access is denied when no rule matches.
	// Defaults to true (deny by default).
	defaultDeny bool

	// denyOverrides controls whether a deny rule overrides an allow rule.
	// Defaults to true.
	denyOverrides bool
}

type OptFunc func(*Options)

// WithRules sets the initial set of ACL rules.
func WithRules(rules []Rule) OptFunc {
	return func(o *Options) { o.rules = rules }
}

// WithRule adds a single allow rule.
func WithRule(subject, action, resource string) OptFunc {
	return func(o *Options) {
		o.rules = append(o.rules, Rule{
			Subject:  subject,
			Action:   action,
			Resource: resource,
			Effect:   "allow",
		})
	}
}

// WithDenyRule adds a single deny rule.
func WithDenyRule(subject, action, resource string) OptFunc {
	return func(o *Options) {
		o.rules = append(o.rules, Rule{
			Subject:  subject,
			Action:   action,
			Resource: resource,
			Effect:   "deny",
		})
	}
}

// WithWildcard sets the wildcard character. Default "*".
func WithWildcard(w string) OptFunc {
	return func(o *Options) { o.wildcard = w }
}

// WithDefaultAllow sets the default to allow when no rule matches.
func WithDefaultAllow() OptFunc {
	return func(o *Options) { o.defaultDeny = false }
}

// WithDenyOverrides controls whether deny rules take precedence.
func WithDenyOverrides(v bool) OptFunc {
	return func(o *Options) { o.denyOverrides = v }
}
