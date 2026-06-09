module github.com/tx7do/go-wind-plugins/config/redis

go 1.26.3

require (
	github.com/redis/go-redis/v9 v9.18.0
	github.com/tx7do/go-wind-plugins/config v0.0.0-00010101000000-000000000000
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	go.uber.org/atomic v1.11.0 // indirect
)

replace github.com/tx7do/go-wind-plugins/config => ../
