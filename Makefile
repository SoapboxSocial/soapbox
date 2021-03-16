.PHONY: protobuf

protobuf:
	 protoc \
 	  --proto_path=$(PROTO_PATH) \
 	  --go_out=plugins=grpc:. \
 	  --grpc-gateway_out=:. \
	  room.proto room_api.proto signal.proto

mock:
	mockgen -package=mocks -destination=pkg/login/internal/mocks/signinwithapple_mock.go -source=pkg/apple/signinwithapple.go
	mockgen -package=mocks -destination=pkg/login/internal/mocks/roomserviceclient_mock.go -source=pkg/rooms/pb/room_api.pb.go RoomServiceClient
.PHONY: mock
