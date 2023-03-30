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

require github.com/google/go-cmp v0.5.9 // indirect
