.PHONY: protobuf

protobuf:
        buf generate https://github.com/soapboxsocial/protobufs.git

mock:
	mockgen -package=mocks -destination=pkg/login/internal/mocks/signinwithapple_mock.go -source=pkg/apple/signinwithapple.go
	mockgen -package=mocks -destination=pkg/login/internal/mocks/roomserviceclient_mock.go -source=pkg/rooms/pb/room_api_grpc.pb.go RoomServiceClient
.PHONY: mock
