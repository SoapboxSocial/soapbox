.PHONY: protobuf

protobuf:
	 protoc --proto_path=$(PROTO_PATH) --go_out=. room.proto