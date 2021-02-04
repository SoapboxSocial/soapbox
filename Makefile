.PHONY: protobuf

protobuf:
	 protoc \
 	  --proto_path=$(PROTO_PATH) \
 	  --go_out=plugins=grpc:. \
 	  --grpc-gateway_out=:. \
	  room.proto room_api.proto signal.proto

mock:
	mockgen -package=internal -destination=pkg/login/internal/signinwithapple_mock.go -source=pkg/apple/signinwithapple.go
.PHONY: mock
