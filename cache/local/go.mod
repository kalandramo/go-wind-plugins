module github.com/tx7do/go-wind-plugins/cache/local

go 1.26.3

require (
	github.com/coocood/freecache v1.2.7
	github.com/tx7do/go-wind-plugins/cache v0.0.1
)

require github.com/cespare/xxhash/v2 v2.3.0 // indirect

replace github.com/tx7do/go-wind-plugins/cache => ../
