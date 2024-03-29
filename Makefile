.PHONY: protobuf

protobuf:
ifdef BRANCH
	buf generate https://github.com/soapboxsocial/protobufs.git#branch=$(BRANCH)
else
	buf generate https://github.com/soapboxsocial/protobufs.git
endif

mock:
	mockgen -package=mocks -destination=mocks/signinwithapple_mock.go -source=pkg/apple/signinwithapple.go
	mockgen -package=mocks -destination=mocks/apns_mock.go -source=pkg/notifications/apns.go
	mockgen -package=mocks -destination=mocks/roomserviceclient_mock.go -source=pkg/rooms/pb/room_api_grpc.pb.go RoomServiceClient
.PHONY: mock

cover:
	go test ./... -coverprofile cover.out
	go tool cover -func cover.out
	rm -f cover.out
.PHONY: cover
