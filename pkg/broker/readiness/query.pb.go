// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.14.0
// source: query.proto

package readiness

import (
	proto "github.com/golang/protobuf/proto"
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

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type GetGenerationReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key string `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"` // namespace/name/UID
}

func (x *GetGenerationReq) Reset() {
	*x = GetGenerationReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_query_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetGenerationReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetGenerationReq) ProtoMessage() {}

func (x *GetGenerationReq) ProtoReflect() protoreflect.Message {
	mi := &file_query_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetGenerationReq.ProtoReflect.Descriptor instead.
func (*GetGenerationReq) Descriptor() ([]byte, []int) {
	return file_query_proto_rawDescGZIP(), []int{0}
}

func (x *GetGenerationReq) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

type GetGenerationResp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Generation int64 `protobuf:"varint,1,opt,name=generation,proto3" json:"generation,omitempty"`
}

func (x *GetGenerationResp) Reset() {
	*x = GetGenerationResp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_query_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetGenerationResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetGenerationResp) ProtoMessage() {}

func (x *GetGenerationResp) ProtoReflect() protoreflect.Message {
	mi := &file_query_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetGenerationResp.ProtoReflect.Descriptor instead.
func (*GetGenerationResp) Descriptor() ([]byte, []int) {
	return file_query_proto_rawDescGZIP(), []int{1}
}

func (x *GetGenerationResp) GetGeneration() int64 {
	if x != nil {
		return x.Generation
	}
	return 0
}

var File_query_proto protoreflect.FileDescriptor

var file_query_proto_rawDesc = []byte{
	0x0a, 0x0b, 0x71, 0x75, 0x65, 0x72, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x09, 0x72,
	0x65, 0x61, 0x64, 0x69, 0x6e, 0x65, 0x73, 0x73, 0x22, 0x24, 0x0a, 0x10, 0x47, 0x65, 0x74, 0x47,
	0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x12, 0x10, 0x0a, 0x03,
	0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x22, 0x33,
	0x0a, 0x11, 0x47, 0x65, 0x74, 0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52,
	0x65, 0x73, 0x70, 0x12, 0x1e, 0x0a, 0x0a, 0x67, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0a, 0x67, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x32, 0xc4, 0x01, 0x0a, 0x16, 0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x51, 0x75, 0x65, 0x72, 0x79, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x56,
	0x0a, 0x17, 0x47, 0x65, 0x74, 0x43, 0x65, 0x6c, 0x6c, 0x54, 0x65, 0x6e, 0x61, 0x6e, 0x74, 0x47,
	0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x1b, 0x2e, 0x72, 0x65, 0x61, 0x64,
	0x69, 0x6e, 0x65, 0x73, 0x73, 0x2e, 0x47, 0x65, 0x74, 0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x1a, 0x1c, 0x2e, 0x72, 0x65, 0x61, 0x64, 0x69, 0x6e, 0x65,
	0x73, 0x73, 0x2e, 0x47, 0x65, 0x74, 0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x52, 0x65, 0x73, 0x70, 0x22, 0x00, 0x12, 0x52, 0x0a, 0x13, 0x47, 0x65, 0x74, 0x54, 0x61, 0x72,
	0x67, 0x65, 0x74, 0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x1b, 0x2e,
	0x72, 0x65, 0x61, 0x64, 0x69, 0x6e, 0x65, 0x73, 0x73, 0x2e, 0x47, 0x65, 0x74, 0x47, 0x65, 0x6e,
	0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x1a, 0x1c, 0x2e, 0x72, 0x65, 0x61,
	0x64, 0x69, 0x6e, 0x65, 0x73, 0x73, 0x2e, 0x47, 0x65, 0x74, 0x47, 0x65, 0x6e, 0x65, 0x72, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x22, 0x00, 0x42, 0x34, 0x5a, 0x32, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f,
	0x6b, 0x6e, 0x61, 0x74, 0x69, 0x76, 0x65, 0x2d, 0x67, 0x63, 0x70, 0x2f, 0x70, 0x6b, 0x67, 0x2f,
	0x62, 0x72, 0x6f, 0x6b, 0x65, 0x72, 0x2f, 0x72, 0x65, 0x61, 0x64, 0x69, 0x6e, 0x65, 0x73, 0x73,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_query_proto_rawDescOnce sync.Once
	file_query_proto_rawDescData = file_query_proto_rawDesc
)

func file_query_proto_rawDescGZIP() []byte {
	file_query_proto_rawDescOnce.Do(func() {
		file_query_proto_rawDescData = protoimpl.X.CompressGZIP(file_query_proto_rawDescData)
	})
	return file_query_proto_rawDescData
}

var file_query_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_query_proto_goTypes = []interface{}{
	(*GetGenerationReq)(nil),  // 0: readiness.GetGenerationReq
	(*GetGenerationResp)(nil), // 1: readiness.GetGenerationResp
}
var file_query_proto_depIdxs = []int32{
	0, // 0: readiness.GenerationQueryService.GetCellTenantGeneration:input_type -> readiness.GetGenerationReq
	0, // 1: readiness.GenerationQueryService.GetTargetGeneration:input_type -> readiness.GetGenerationReq
	1, // 2: readiness.GenerationQueryService.GetCellTenantGeneration:output_type -> readiness.GetGenerationResp
	1, // 3: readiness.GenerationQueryService.GetTargetGeneration:output_type -> readiness.GetGenerationResp
	2, // [2:4] is the sub-list for method output_type
	0, // [0:2] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_query_proto_init() }
func file_query_proto_init() {
	if File_query_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_query_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetGenerationReq); i {
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
		file_query_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetGenerationResp); i {
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
			RawDescriptor: file_query_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_query_proto_goTypes,
		DependencyIndexes: file_query_proto_depIdxs,
		MessageInfos:      file_query_proto_msgTypes,
	}.Build()
	File_query_proto = out.File
	file_query_proto_rawDesc = nil
	file_query_proto_goTypes = nil
	file_query_proto_depIdxs = nil
}
