.PHONY: protobuf

protobuf:
ifdef BRANCH
	buf generate https://github.com/soapboxsocial/protobufs.git#branch=$(BRANCH)
else
	buf generate https://github.com/soapboxsocial/protobufs.git
endif

mock:
	mockgen -package=mocks -destination=pkg/login/internal/mocks/signinwithapple_mock.go -source=pkg/apple/signinwithapple.go
	mockgen -package=mocks -destination=pkg/login/internal/mocks/roomserviceclient_mock.go -source=pkg/rooms/pb/room_api_grpc.pb.go RoomServiceClient
	mockgen -package=mocks -destination=pkg/notifications/builders/internal/mocks/roomserviceclient_mock.go -source=pkg/rooms/pb/room_api_grpc.pb.go RoomServiceClient
.PHONY: mock
