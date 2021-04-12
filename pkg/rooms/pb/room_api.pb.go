// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.15.2
// source: soapbox/v1/room_api.proto

package pb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type GetRoomRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *GetRoomRequest) Reset() {
	*x = GetRoomRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_soapbox_v1_room_api_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetRoomRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetRoomRequest) ProtoMessage() {}

func (x *GetRoomRequest) ProtoReflect() protoreflect.Message {
	mi := &file_soapbox_v1_room_api_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetRoomRequest.ProtoReflect.Descriptor instead.
func (*GetRoomRequest) Descriptor() ([]byte, []int) {
	return file_soapbox_v1_room_api_proto_rawDescGZIP(), []int{0}
}

func (x *GetRoomRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type GetRoomResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	State *RoomState `protobuf:"bytes,1,opt,name=state,proto3" json:"state,omitempty"`
}

func (x *GetRoomResponse) Reset() {
	*x = GetRoomResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_soapbox_v1_room_api_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetRoomResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetRoomResponse) ProtoMessage() {}

func (x *GetRoomResponse) ProtoReflect() protoreflect.Message {
	mi := &file_soapbox_v1_room_api_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetRoomResponse.ProtoReflect.Descriptor instead.
func (*GetRoomResponse) Descriptor() ([]byte, []int) {
	return file_soapbox_v1_room_api_proto_rawDescGZIP(), []int{1}
}

func (x *GetRoomResponse) GetState() *RoomState {
	if x != nil {
		return x.State
	}
	return nil
}

type ListRoomsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *ListRoomsRequest) Reset() {
	*x = ListRoomsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_soapbox_v1_room_api_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListRoomsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListRoomsRequest) ProtoMessage() {}

func (x *ListRoomsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_soapbox_v1_room_api_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListRoomsRequest.ProtoReflect.Descriptor instead.
func (*ListRoomsRequest) Descriptor() ([]byte, []int) {
	return file_soapbox_v1_room_api_proto_rawDescGZIP(), []int{2}
}

type ListRoomsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Rooms []*RoomState `protobuf:"bytes,1,rep,name=rooms,proto3" json:"rooms,omitempty"`
}

func (x *ListRoomsResponse) Reset() {
	*x = ListRoomsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_soapbox_v1_room_api_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListRoomsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListRoomsResponse) ProtoMessage() {}

func (x *ListRoomsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_soapbox_v1_room_api_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListRoomsResponse.ProtoReflect.Descriptor instead.
func (*ListRoomsResponse) Descriptor() ([]byte, []int) {
	return file_soapbox_v1_room_api_proto_rawDescGZIP(), []int{3}
}

func (x *ListRoomsResponse) GetRooms() []*RoomState {
	if x != nil {
		return x.Rooms
	}
	return nil
}

type CloseRoomRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *CloseRoomRequest) Reset() {
	*x = CloseRoomRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_soapbox_v1_room_api_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CloseRoomRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CloseRoomRequest) ProtoMessage() {}

func (x *CloseRoomRequest) ProtoReflect() protoreflect.Message {
	mi := &file_soapbox_v1_room_api_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CloseRoomRequest.ProtoReflect.Descriptor instead.
func (*CloseRoomRequest) Descriptor() ([]byte, []int) {
	return file_soapbox_v1_room_api_proto_rawDescGZIP(), []int{4}
}

func (x *CloseRoomRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type CloseRoomResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Success bool `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
}

func (x *CloseRoomResponse) Reset() {
	*x = CloseRoomResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_soapbox_v1_room_api_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CloseRoomResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CloseRoomResponse) ProtoMessage() {}

func (x *CloseRoomResponse) ProtoReflect() protoreflect.Message {
	mi := &file_soapbox_v1_room_api_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CloseRoomResponse.ProtoReflect.Descriptor instead.
func (*CloseRoomResponse) Descriptor() ([]byte, []int) {
	return file_soapbox_v1_room_api_proto_rawDescGZIP(), []int{5}
}

func (x *CloseRoomResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

type RegisterWelcomeRoomRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	UserId int64 `protobuf:"varint,1,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
}

func (x *RegisterWelcomeRoomRequest) Reset() {
	*x = RegisterWelcomeRoomRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_soapbox_v1_room_api_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RegisterWelcomeRoomRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RegisterWelcomeRoomRequest) ProtoMessage() {}

func (x *RegisterWelcomeRoomRequest) ProtoReflect() protoreflect.Message {
	mi := &file_soapbox_v1_room_api_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RegisterWelcomeRoomRequest.ProtoReflect.Descriptor instead.
func (*RegisterWelcomeRoomRequest) Descriptor() ([]byte, []int) {
	return file_soapbox_v1_room_api_proto_rawDescGZIP(), []int{6}
}

func (x *RegisterWelcomeRoomRequest) GetUserId() int64 {
	if x != nil {
		return x.UserId
	}
	return 0
}

type RegisterWelcomeRoomResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *RegisterWelcomeRoomResponse) Reset() {
	*x = RegisterWelcomeRoomResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_soapbox_v1_room_api_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RegisterWelcomeRoomResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RegisterWelcomeRoomResponse) ProtoMessage() {}

func (x *RegisterWelcomeRoomResponse) ProtoReflect() protoreflect.Message {
	mi := &file_soapbox_v1_room_api_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RegisterWelcomeRoomResponse.ProtoReflect.Descriptor instead.
func (*RegisterWelcomeRoomResponse) Descriptor() ([]byte, []int) {
	return file_soapbox_v1_room_api_proto_rawDescGZIP(), []int{7}
}

func (x *RegisterWelcomeRoomResponse) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

var File_soapbox_v1_room_api_proto protoreflect.FileDescriptor

var file_soapbox_v1_room_api_proto_rawDesc = []byte{
	0x0a, 0x19, 0x73, 0x6f, 0x61, 0x70, 0x62, 0x6f, 0x78, 0x2f, 0x76, 0x31, 0x2f, 0x72, 0x6f, 0x6f,
	0x6d, 0x5f, 0x61, 0x70, 0x69, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0a, 0x73, 0x6f, 0x61,
	0x70, 0x62, 0x6f, 0x78, 0x2e, 0x76, 0x31, 0x1a, 0x15, 0x73, 0x6f, 0x61, 0x70, 0x62, 0x6f, 0x78,
	0x2f, 0x76, 0x31, 0x2f, 0x72, 0x6f, 0x6f, 0x6d, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x20,
	0x0a, 0x0e, 0x47, 0x65, 0x74, 0x52, 0x6f, 0x6f, 0x6d, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64,
	0x22, 0x3e, 0x0a, 0x0f, 0x47, 0x65, 0x74, 0x52, 0x6f, 0x6f, 0x6d, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x2b, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x15, 0x2e, 0x73, 0x6f, 0x61, 0x70, 0x62, 0x6f, 0x78, 0x2e, 0x76, 0x31, 0x2e,
	0x52, 0x6f, 0x6f, 0x6d, 0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65,
	0x22, 0x12, 0x0a, 0x10, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x6f, 0x6f, 0x6d, 0x73, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x22, 0x40, 0x0a, 0x11, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x6f, 0x6f, 0x6d,
	0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2b, 0x0a, 0x05, 0x72, 0x6f, 0x6f,
	0x6d, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x73, 0x6f, 0x61, 0x70, 0x62,
	0x6f, 0x78, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x6f, 0x6f, 0x6d, 0x53, 0x74, 0x61, 0x74, 0x65, 0x52,
	0x05, 0x72, 0x6f, 0x6f, 0x6d, 0x73, 0x22, 0x22, 0x0a, 0x10, 0x43, 0x6c, 0x6f, 0x73, 0x65, 0x52,
	0x6f, 0x6f, 0x6d, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x22, 0x2d, 0x0a, 0x11, 0x43, 0x6c,
	0x6f, 0x73, 0x65, 0x52, 0x6f, 0x6f, 0x6d, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x18, 0x0a, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08,
	0x52, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x22, 0x35, 0x0a, 0x1a, 0x52, 0x65, 0x67,
	0x69, 0x73, 0x74, 0x65, 0x72, 0x57, 0x65, 0x6c, 0x63, 0x6f, 0x6d, 0x65, 0x52, 0x6f, 0x6f, 0x6d,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x17, 0x0a, 0x07, 0x75, 0x73, 0x65, 0x72, 0x5f,
	0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x75, 0x73, 0x65, 0x72, 0x49, 0x64,
	0x22, 0x2d, 0x0a, 0x1b, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x57, 0x65, 0x6c, 0x63,
	0x6f, 0x6d, 0x65, 0x52, 0x6f, 0x6f, 0x6d, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x32,
	0xcd, 0x02, 0x0a, 0x0b, 0x52, 0x6f, 0x6f, 0x6d, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12,
	0x42, 0x0a, 0x07, 0x47, 0x65, 0x74, 0x52, 0x6f, 0x6f, 0x6d, 0x12, 0x1a, 0x2e, 0x73, 0x6f, 0x61,
	0x70, 0x62, 0x6f, 0x78, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x52, 0x6f, 0x6f, 0x6d, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1b, 0x2e, 0x73, 0x6f, 0x61, 0x70, 0x62, 0x6f, 0x78,
	0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x52, 0x6f, 0x6f, 0x6d, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x48, 0x0a, 0x09, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x6f, 0x6f, 0x6d, 0x73,
	0x12, 0x1c, 0x2e, 0x73, 0x6f, 0x61, 0x70, 0x62, 0x6f, 0x78, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x69,
	0x73, 0x74, 0x52, 0x6f, 0x6f, 0x6d, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1d,
	0x2e, 0x73, 0x6f, 0x61, 0x70, 0x62, 0x6f, 0x78, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x69, 0x73, 0x74,
	0x52, 0x6f, 0x6f, 0x6d, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x48, 0x0a,
	0x09, 0x43, 0x6c, 0x6f, 0x73, 0x65, 0x52, 0x6f, 0x6f, 0x6d, 0x12, 0x1c, 0x2e, 0x73, 0x6f, 0x61,
	0x70, 0x62, 0x6f, 0x78, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x6c, 0x6f, 0x73, 0x65, 0x52, 0x6f, 0x6f,
	0x6d, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1d, 0x2e, 0x73, 0x6f, 0x61, 0x70, 0x62,
	0x6f, 0x78, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x6c, 0x6f, 0x73, 0x65, 0x52, 0x6f, 0x6f, 0x6d, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x66, 0x0a, 0x13, 0x52, 0x65, 0x67, 0x69, 0x73,
	0x74, 0x65, 0x72, 0x57, 0x65, 0x6c, 0x63, 0x6f, 0x6d, 0x65, 0x52, 0x6f, 0x6f, 0x6d, 0x12, 0x26,
	0x2e, 0x73, 0x6f, 0x61, 0x70, 0x62, 0x6f, 0x78, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x67, 0x69,
	0x73, 0x74, 0x65, 0x72, 0x57, 0x65, 0x6c, 0x63, 0x6f, 0x6d, 0x65, 0x52, 0x6f, 0x6f, 0x6d, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x27, 0x2e, 0x73, 0x6f, 0x61, 0x70, 0x62, 0x6f, 0x78,
	0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x57, 0x65, 0x6c, 0x63,
	0x6f, 0x6d, 0x65, 0x52, 0x6f, 0x6f, 0x6d, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42,
	0x0e, 0x5a, 0x0c, 0x70, 0x6b, 0x67, 0x2f, 0x72, 0x6f, 0x6f, 0x6d, 0x73, 0x2f, 0x70, 0x62, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_soapbox_v1_room_api_proto_rawDescOnce sync.Once
	file_soapbox_v1_room_api_proto_rawDescData = file_soapbox_v1_room_api_proto_rawDesc
)

func file_soapbox_v1_room_api_proto_rawDescGZIP() []byte {
	file_soapbox_v1_room_api_proto_rawDescOnce.Do(func() {
		file_soapbox_v1_room_api_proto_rawDescData = protoimpl.X.CompressGZIP(file_soapbox_v1_room_api_proto_rawDescData)
	})
	return file_soapbox_v1_room_api_proto_rawDescData
}

var file_soapbox_v1_room_api_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_soapbox_v1_room_api_proto_goTypes = []interface{}{
	(*GetRoomRequest)(nil),              // 0: soapbox.v1.GetRoomRequest
	(*GetRoomResponse)(nil),             // 1: soapbox.v1.GetRoomResponse
	(*ListRoomsRequest)(nil),            // 2: soapbox.v1.ListRoomsRequest
	(*ListRoomsResponse)(nil),           // 3: soapbox.v1.ListRoomsResponse
	(*CloseRoomRequest)(nil),            // 4: soapbox.v1.CloseRoomRequest
	(*CloseRoomResponse)(nil),           // 5: soapbox.v1.CloseRoomResponse
	(*RegisterWelcomeRoomRequest)(nil),  // 6: soapbox.v1.RegisterWelcomeRoomRequest
	(*RegisterWelcomeRoomResponse)(nil), // 7: soapbox.v1.RegisterWelcomeRoomResponse
	(*RoomState)(nil),                   // 8: soapbox.v1.RoomState
}
var file_soapbox_v1_room_api_proto_depIdxs = []int32{
	8, // 0: soapbox.v1.GetRoomResponse.state:type_name -> soapbox.v1.RoomState
	8, // 1: soapbox.v1.ListRoomsResponse.rooms:type_name -> soapbox.v1.RoomState
	0, // 2: soapbox.v1.RoomService.GetRoom:input_type -> soapbox.v1.GetRoomRequest
	2, // 3: soapbox.v1.RoomService.ListRooms:input_type -> soapbox.v1.ListRoomsRequest
	4, // 4: soapbox.v1.RoomService.CloseRoom:input_type -> soapbox.v1.CloseRoomRequest
	6, // 5: soapbox.v1.RoomService.RegisterWelcomeRoom:input_type -> soapbox.v1.RegisterWelcomeRoomRequest
	1, // 6: soapbox.v1.RoomService.GetRoom:output_type -> soapbox.v1.GetRoomResponse
	3, // 7: soapbox.v1.RoomService.ListRooms:output_type -> soapbox.v1.ListRoomsResponse
	5, // 8: soapbox.v1.RoomService.CloseRoom:output_type -> soapbox.v1.CloseRoomResponse
	7, // 9: soapbox.v1.RoomService.RegisterWelcomeRoom:output_type -> soapbox.v1.RegisterWelcomeRoomResponse
	6, // [6:10] is the sub-list for method output_type
	2, // [2:6] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_soapbox_v1_room_api_proto_init() }
func file_soapbox_v1_room_api_proto_init() {
	if File_soapbox_v1_room_api_proto != nil {
		return
	}
	file_soapbox_v1_room_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_soapbox_v1_room_api_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetRoomRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_soapbox_v1_room_api_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetRoomResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_soapbox_v1_room_api_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListRoomsRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_soapbox_v1_room_api_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListRoomsResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_soapbox_v1_room_api_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CloseRoomRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_soapbox_v1_room_api_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CloseRoomResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_soapbox_v1_room_api_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RegisterWelcomeRoomRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_soapbox_v1_room_api_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RegisterWelcomeRoomResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_soapbox_v1_room_api_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_soapbox_v1_room_api_proto_goTypes,
		DependencyIndexes: file_soapbox_v1_room_api_proto_depIdxs,
		MessageInfos:      file_soapbox_v1_room_api_proto_msgTypes,
	}.Build()
	File_soapbox_v1_room_api_proto = out.File
	file_soapbox_v1_room_api_proto_rawDesc = nil
	file_soapbox_v1_room_api_proto_goTypes = nil
	file_soapbox_v1_room_api_proto_depIdxs = nil
}
