.PHONY: protobuf

protobuf:
	 protoc \
 	  --proto_path=$(PROTO_PATH) \
 	  --go_out=plugins=grpc:. \
 	  --grpc-gateway_out=:. \
 	  room.proto signal.proto
