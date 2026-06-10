module github.com/tx7do/go-wind-plugins/transport/socketio

go 1.26.3

require (
	github.com/googollee/go-socket.io v1.7.0
	github.com/gorilla/handlers v1.5.2
	github.com/gorilla/mux v1.8.1
	github.com/tx7do/go-wind v0.0.1
	github.com/tx7do/go-wind-plugins/encoding v0.0.1
)

require (
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/gofrs/uuid v4.4.0+incompatible // indirect
	github.com/gomodule/redigo v1.9.3 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
)

replace github.com/tx7do/go-wind-plugins/encoding => ../../encoding
