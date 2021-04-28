// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.15.2
// source: soapbox/notifications/v1/notifications.proto

package pb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	anypb "google.golang.org/protobuf/types/known/anypb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Notification struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Targets []int64            `protobuf:"varint,1,rep,packed,name=targets,proto3" json:"targets,omitempty"`
	Data    *Notification_Data `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *Notification) Reset() {
	*x = Notification{}
	if protoimpl.UnsafeEnabled {
		mi := &file_soapbox_notifications_v1_notifications_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Notification) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Notification) ProtoMessage() {}

func (x *Notification) ProtoReflect() protoreflect.Message {
	mi := &file_soapbox_notifications_v1_notifications_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Notification.ProtoReflect.Descriptor instead.
func (*Notification) Descriptor() ([]byte, []int) {
	return file_soapbox_notifications_v1_notifications_proto_rawDescGZIP(), []int{0}
}

func (x *Notification) GetTargets() []int64 {
	if x != nil {
		return x.Targets
	}
	return nil
}

func (x *Notification) GetData() *Notification_Data {
	if x != nil {
		return x.Data
	}
	return nil
}

type Notification_Data struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Category              string                `protobuf:"bytes,1,opt,name=category,proto3" json:"category,omitempty"`
	LocalizationArguments []string              `protobuf:"bytes,2,rep,name=localization_arguments,json=localizationArguments,proto3" json:"localization_arguments,omitempty"`
	Projects              map[string]*anypb.Any `protobuf:"bytes,3,rep,name=projects,proto3" json:"projects,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *Notification_Data) Reset() {
	*x = Notification_Data{}
	if protoimpl.UnsafeEnabled {
		mi := &file_soapbox_notifications_v1_notifications_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Notification_Data) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Notification_Data) ProtoMessage() {}

func (x *Notification_Data) ProtoReflect() protoreflect.Message {
	mi := &file_soapbox_notifications_v1_notifications_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Notification_Data.ProtoReflect.Descriptor instead.
func (*Notification_Data) Descriptor() ([]byte, []int) {
	return file_soapbox_notifications_v1_notifications_proto_rawDescGZIP(), []int{0, 0}
}

func (x *Notification_Data) GetCategory() string {
	if x != nil {
		return x.Category
	}
	return ""
}

func (x *Notification_Data) GetLocalizationArguments() []string {
	if x != nil {
		return x.LocalizationArguments
	}
	return nil
}

func (x *Notification_Data) GetProjects() map[string]*anypb.Any {
	if x != nil {
		return x.Projects
	}
	return nil
}

var File_soapbox_notifications_v1_notifications_proto protoreflect.FileDescriptor

var file_soapbox_notifications_v1_notifications_proto_rawDesc = []byte{
	0x0a, 0x2c, 0x73, 0x6f, 0x61, 0x70, 0x62, 0x6f, 0x78, 0x2f, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x69,
	0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2f, 0x76, 0x31, 0x2f, 0x6e, 0x6f, 0x74, 0x69, 0x66,
	0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x18,
	0x73, 0x6f, 0x61, 0x70, 0x62, 0x6f, 0x78, 0x2e, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x76, 0x31, 0x1a, 0x19, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x61, 0x6e, 0x79, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0xef, 0x02, 0x0a, 0x0c, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x12, 0x18, 0x0a, 0x07, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x73, 0x18,
	0x01, 0x20, 0x03, 0x28, 0x03, 0x52, 0x07, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x73, 0x12, 0x3f,
	0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2b, 0x2e, 0x73,
	0x6f, 0x61, 0x70, 0x62, 0x6f, 0x78, 0x2e, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x44, 0x61, 0x74, 0x61, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x1a,
	0x83, 0x02, 0x0a, 0x04, 0x44, 0x61, 0x74, 0x61, 0x12, 0x1a, 0x0a, 0x08, 0x63, 0x61, 0x74, 0x65,
	0x67, 0x6f, 0x72, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x63, 0x61, 0x74, 0x65,
	0x67, 0x6f, 0x72, 0x79, 0x12, 0x35, 0x0a, 0x16, 0x6c, 0x6f, 0x63, 0x61, 0x6c, 0x69, 0x7a, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x61, 0x72, 0x67, 0x75, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x18, 0x02,
	0x20, 0x03, 0x28, 0x09, 0x52, 0x15, 0x6c, 0x6f, 0x63, 0x61, 0x6c, 0x69, 0x7a, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x41, 0x72, 0x67, 0x75, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x12, 0x55, 0x0a, 0x08, 0x70,
	0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x39, 0x2e,
	0x73, 0x6f, 0x61, 0x70, 0x62, 0x6f, 0x78, 0x2e, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x76, 0x31, 0x2e, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x44, 0x61, 0x74, 0x61, 0x2e, 0x50, 0x72, 0x6f, 0x6a, 0x65,
	0x63, 0x74, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x08, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63,
	0x74, 0x73, 0x1a, 0x51, 0x0a, 0x0d, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x73, 0x45, 0x6e,
	0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x2a, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x41, 0x6e, 0x79, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x3a, 0x02, 0x38, 0x01, 0x42, 0x16, 0x5a, 0x14, 0x70, 0x6b, 0x67, 0x2f, 0x6e, 0x6f, 0x74,
	0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2f, 0x70, 0x62, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_soapbox_notifications_v1_notifications_proto_rawDescOnce sync.Once
	file_soapbox_notifications_v1_notifications_proto_rawDescData = file_soapbox_notifications_v1_notifications_proto_rawDesc
)

func file_soapbox_notifications_v1_notifications_proto_rawDescGZIP() []byte {
	file_soapbox_notifications_v1_notifications_proto_rawDescOnce.Do(func() {
		file_soapbox_notifications_v1_notifications_proto_rawDescData = protoimpl.X.CompressGZIP(file_soapbox_notifications_v1_notifications_proto_rawDescData)
	})
	return file_soapbox_notifications_v1_notifications_proto_rawDescData
}

var file_soapbox_notifications_v1_notifications_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_soapbox_notifications_v1_notifications_proto_goTypes = []interface{}{
	(*Notification)(nil),      // 0: soapbox.notifications.v1.Notification
	(*Notification_Data)(nil), // 1: soapbox.notifications.v1.Notification.Data
	nil,                       // 2: soapbox.notifications.v1.Notification.Data.ProjectsEntry
	(*anypb.Any)(nil),         // 3: google.protobuf.Any
}
var file_soapbox_notifications_v1_notifications_proto_depIdxs = []int32{
	1, // 0: soapbox.notifications.v1.Notification.data:type_name -> soapbox.notifications.v1.Notification.Data
	2, // 1: soapbox.notifications.v1.Notification.Data.projects:type_name -> soapbox.notifications.v1.Notification.Data.ProjectsEntry
	3, // 2: soapbox.notifications.v1.Notification.Data.ProjectsEntry.value:type_name -> google.protobuf.Any
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_soapbox_notifications_v1_notifications_proto_init() }
func file_soapbox_notifications_v1_notifications_proto_init() {
	if File_soapbox_notifications_v1_notifications_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_soapbox_notifications_v1_notifications_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Notification); i {
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
		file_soapbox_notifications_v1_notifications_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Notification_Data); i {
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
			RawDescriptor: file_soapbox_notifications_v1_notifications_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_soapbox_notifications_v1_notifications_proto_goTypes,
		DependencyIndexes: file_soapbox_notifications_v1_notifications_proto_depIdxs,
		MessageInfos:      file_soapbox_notifications_v1_notifications_proto_msgTypes,
	}.Build()
	File_soapbox_notifications_v1_notifications_proto = out.File
	file_soapbox_notifications_v1_notifications_proto_rawDesc = nil
	file_soapbox_notifications_v1_notifications_proto_goTypes = nil
	file_soapbox_notifications_v1_notifications_proto_depIdxs = nil
}
