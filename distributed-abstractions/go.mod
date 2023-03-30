module hw

go 1.20

replace hw/protobuf => ../protobuf

replace hw/log => ../log

replace hw/state => ../state

replace hw/abstraction => ../abstraction

replace hw/queue => ../queue

replace hw/qprocessor => ../qprocessor

require (
	github.com/google/uuid v1.3.0
	google.golang.org/protobuf v1.30.0
)

require (
	github.com/Ozoniuss/stdlog v1.0.0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.24.0 // indirect
	golang.org/x/net v0.5.0 // indirect
	golang.org/x/sys v0.4.0 // indirect
	golang.org/x/text v0.6.0 // indirect
	google.golang.org/genproto v0.0.0-20230110181048-76db0878b65f // indirect
	google.golang.org/grpc v1.53.0 // indirect
)
