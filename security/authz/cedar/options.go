package cedar

import (
	"net/http"
	"time"
)

// Options holds configuration for the Cedar (AWS Verified Permissions) engine.
type Options struct {
	// policyStoreID is the AWS Verified Permissions policy store ID.
	policyStoreID string

	// endpoint is the AWS Verified Permissions endpoint URL.
	// If set, overrides the default regional endpoint.
	endpoint string

	// region is the AWS region (e.g. "us-east-1").
	region string

	// httpClient allows the caller to inject a custom HTTP client.
	httpClient *http.Client

	// authorizer is an optional callback for evaluating authorization
	// decisions locally (e.g. when using a custom Cedar evaluator or
	// for testing). When set, it replaces the remote AVP call.
	authorizer AuthorizerFunc
}

// AuthorizerFunc is the signature for a local authorization callback.
type AuthorizerFunc func(principal, action, resource string) (bool, error)

type OptFunc func(*Options)

// WithPolicyStoreID sets the AWS Verified Permissions policy store ID.
func WithPolicyStoreID(id string) OptFunc {
	return func(o *Options) { o.policyStoreID = id }
}

// WithEndpoint overrides the AWS VP endpoint URL.
func WithEndpoint(url string) OptFunc {
	return func(o *Options) { o.endpoint = url }
}

// WithRegion sets the AWS region.
func WithRegion(region string) OptFunc {
	return func(o *Options) { o.region = region }
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(c *http.Client) OptFunc {
	return func(o *Options) { o.httpClient = c }
}

// WithAuthorizer sets a local authorization callback for testing or
// custom Cedar evaluation.
func WithAuthorizer(fn AuthorizerFunc) OptFunc {
	return func(o *Options) { o.authorizer = fn }
}

func (o *Options) getHTTPClient() *http.Client {
	if o.httpClient != nil {
		return o.httpClient
	}
	return &http.Client{Timeout: 10 * time.Second}
}

func (o *Options) getEndpoint() string {
	if o.endpoint != "" {
		return o.endpoint
	}
	region := o.region
	if region == "" {
		region = "us-east-1"
	}
	return "https://verifiedpermissions." + region + ".amazonaws.com"
}
