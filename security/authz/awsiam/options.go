package awsiam

// Statement represents a single IAM policy statement.
type Statement struct {
	Effect    string   `json:"Effect"`
	Actions   []string `json:"Action"`
	Resources []string `json:"Resource"`
}

// Policy represents an AWS IAM policy document.
type Policy struct {
	Version   string      `json:"Version"`
	Statement []Statement `json:"Statement"`
}

// Options holds configuration for the AWS IAM engine.
type Options struct {
	// policies maps subject → list of policies.
	policies map[string][]Policy

	// defaultDeny controls whether access is denied when no policy matches.
	// Defaults to true.
	defaultDeny bool
}

type OptFunc func(*Options)

// WithPolicy attaches an IAM policy to a subject.
func WithPolicy(subject string, policy Policy) OptFunc {
	return func(o *Options) {
		if o.policies == nil {
			o.policies = make(map[string][]Policy)
		}
		o.policies[subject] = append(o.policies[subject], policy)
	}
}

// WithPolicies sets the entire subject→policies map.
func WithPolicies(policies map[string][]Policy) OptFunc {
	return func(o *Options) { o.policies = policies }
}

// WithAllowStatement adds a simple Allow statement to a subject.
func WithAllowStatement(subject string, actions, resources []string) OptFunc {
	return func(o *Options) {
		if o.policies == nil {
			o.policies = make(map[string][]Policy)
		}
		o.policies[subject] = append(o.policies[subject], Policy{
			Version: "2012-10-17",
			Statement: []Statement{{
				Effect:    "Allow",
				Actions:   actions,
				Resources: resources,
			}},
		})
	}
}

// WithDenyStatement adds a simple Deny statement to a subject.
func WithDenyStatement(subject string, actions, resources []string) OptFunc {
	return func(o *Options) {
		if o.policies == nil {
			o.policies = make(map[string][]Policy)
		}
		o.policies[subject] = append(o.policies[subject], Policy{
			Version: "2012-10-17",
			Statement: []Statement{{
				Effect:    "Deny",
				Actions:   resources,
				Resources: resources,
			}},
		})
	}
}

// WithDefaultAllow sets the default to allow when no policy matches.
func WithDefaultAllow() OptFunc {
	return func(o *Options) { o.defaultDeny = false }
}
