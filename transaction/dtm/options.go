package dtm

// Option configures the DTM client.
type Option func(*config)

// WithServer sets the DTM server address.
// Default: "http://localhost:36789/api/dtmsvr".
func WithServer(server string) Option {
	return func(c *config) {
		c.Server = server
	}
}
