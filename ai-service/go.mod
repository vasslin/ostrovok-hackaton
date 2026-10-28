module github.com/godnota/ai-service

go 1.26.0

require (
	github.com/caarlos0/env/v10 v10.0.0
	github.com/godnota/pkg v0.0.0
	google.golang.org/grpc v1.84.0
)

require (
	golang.org/x/net v0.58.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
	golang.org/x/text v0.42.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260706201446-f0a921348800 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
)

replace github.com/godnota/pkg => ../pkg
