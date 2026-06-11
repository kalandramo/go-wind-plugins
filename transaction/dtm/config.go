package dtm

const defaultServer = "http://localhost:36789/api/dtmsvr"

// config holds the DTM client configuration.
type config struct {
	// Server is the DTM server address (e.g. "http://localhost:36789/api/dtmsvr").
	Server string
}
