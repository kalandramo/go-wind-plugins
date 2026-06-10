module github.com/tx7do/go-wind-plugins/transport/sse

go 1.26.3

require (
	github.com/google/uuid v1.6.0
	github.com/tx7do/go-wind v0.0.1
	github.com/tx7do/go-wind-plugins/encoding v0.0.1
	github.com/tx7do/go-wind-plugins/encoding/json v0.0.1
)

replace (
	github.com/tx7do/go-wind-plugins/encoding => ../../encoding
	github.com/tx7do/go-wind-plugins/encoding/json => ../../encoding/json
)
