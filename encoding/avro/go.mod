module github.com/tx7do/go-wind-plugins/encoding/avro

go 1.26.3

require (
	github.com/linkedin/goavro/v2 v2.15.0
	github.com/tx7do/go-wind-plugins/encoding v0.0.1
)

require github.com/golang/snappy v1.0.0 // indirect

replace github.com/tx7do/go-wind-plugins/encoding => ../
