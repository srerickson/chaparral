// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        (unknown)
// source: chaparral/v1/manage_service.proto

package chaparralv1

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

// StreamObjectRootsRequest is used to scan an OCFL storage root for objects.
type StreamObjectRootsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	StorageRootId string `protobuf:"bytes,1,opt,name=storage_root_id,json=storageRootId,proto3" json:"storage_root_id,omitempty"`
}

func (x *StreamObjectRootsRequest) Reset() {
	*x = StreamObjectRootsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chaparral_v1_manage_service_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StreamObjectRootsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StreamObjectRootsRequest) ProtoMessage() {}

func (x *StreamObjectRootsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_chaparral_v1_manage_service_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StreamObjectRootsRequest.ProtoReflect.Descriptor instead.
func (*StreamObjectRootsRequest) Descriptor() ([]byte, []int) {
	return file_chaparral_v1_manage_service_proto_rawDescGZIP(), []int{0}
}

func (x *StreamObjectRootsRequest) GetStorageRootId() string {
	if x != nil {
		return x.StorageRootId
	}
	return ""
}

// StreamObjectRootsResponse is used to return OCFL object information during a
// scan of an OCFL storage root.i
type StreamObjectRootsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// path of the object root OCFL object.
	ObjectPath string `protobuf:"bytes,1,opt,name=object_path,json=objectPath,proto3" json:"object_path,omitempty"`
	// the OCFL spec declarated in the objec root
	Spec string `protobuf:"bytes,2,opt,name=spec,proto3" json:"spec,omitempty"`
	// The digest algorithm used for inventory sidecar
	DigestAlgorithm string `protobuf:"bytes,3,opt,name=digest_algorithm,json=digestAlgorithm,proto3" json:"digest_algorithm,omitempty"`
}

func (x *StreamObjectRootsResponse) Reset() {
	*x = StreamObjectRootsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chaparral_v1_manage_service_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StreamObjectRootsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StreamObjectRootsResponse) ProtoMessage() {}

func (x *StreamObjectRootsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_chaparral_v1_manage_service_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StreamObjectRootsResponse.ProtoReflect.Descriptor instead.
func (*StreamObjectRootsResponse) Descriptor() ([]byte, []int) {
	return file_chaparral_v1_manage_service_proto_rawDescGZIP(), []int{1}
}

func (x *StreamObjectRootsResponse) GetObjectPath() string {
	if x != nil {
		return x.ObjectPath
	}
	return ""
}

func (x *StreamObjectRootsResponse) GetSpec() string {
	if x != nil {
		return x.Spec
	}
	return ""
}

func (x *StreamObjectRootsResponse) GetDigestAlgorithm() string {
	if x != nil {
		return x.DigestAlgorithm
	}
	return ""
}

var File_chaparral_v1_manage_service_proto protoreflect.FileDescriptor

var file_chaparral_v1_manage_service_proto_rawDesc = []byte{
	0x0a, 0x21, 0x63, 0x68, 0x61, 0x70, 0x61, 0x72, 0x72, 0x61, 0x6c, 0x2f, 0x76, 0x31, 0x2f, 0x6d,
	0x61, 0x6e, 0x61, 0x67, 0x65, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x0c, 0x63, 0x68, 0x61, 0x70, 0x61, 0x72, 0x72, 0x61, 0x6c, 0x2e, 0x76,
	0x31, 0x22, 0x42, 0x0a, 0x18, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x4f, 0x62, 0x6a, 0x65, 0x63,
	0x74, 0x52, 0x6f, 0x6f, 0x74, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x26, 0x0a,
	0x0f, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x5f, 0x72, 0x6f, 0x6f, 0x74, 0x5f, 0x69, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x52,
	0x6f, 0x6f, 0x74, 0x49, 0x64, 0x22, 0x7b, 0x0a, 0x19, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x4f,
	0x62, 0x6a, 0x65, 0x63, 0x74, 0x52, 0x6f, 0x6f, 0x74, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x1f, 0x0a, 0x0b, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x5f, 0x70, 0x61, 0x74,
	0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x50,
	0x61, 0x74, 0x68, 0x12, 0x12, 0x0a, 0x04, 0x73, 0x70, 0x65, 0x63, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x04, 0x73, 0x70, 0x65, 0x63, 0x12, 0x29, 0x0a, 0x10, 0x64, 0x69, 0x67, 0x65, 0x73,
	0x74, 0x5f, 0x61, 0x6c, 0x67, 0x6f, 0x72, 0x69, 0x74, 0x68, 0x6d, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0f, 0x64, 0x69, 0x67, 0x65, 0x73, 0x74, 0x41, 0x6c, 0x67, 0x6f, 0x72, 0x69, 0x74,
	0x68, 0x6d, 0x32, 0x79, 0x0a, 0x0d, 0x4d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x53, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x12, 0x68, 0x0a, 0x11, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x4f, 0x62, 0x6a,
	0x65, 0x63, 0x74, 0x52, 0x6f, 0x6f, 0x74, 0x73, 0x12, 0x26, 0x2e, 0x63, 0x68, 0x61, 0x70, 0x61,
	0x72, 0x72, 0x61, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x4f, 0x62,
	0x6a, 0x65, 0x63, 0x74, 0x52, 0x6f, 0x6f, 0x74, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x27, 0x2e, 0x63, 0x68, 0x61, 0x70, 0x61, 0x72, 0x72, 0x61, 0x6c, 0x2e, 0x76, 0x31, 0x2e,
	0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x52, 0x6f, 0x6f, 0x74,
	0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x30, 0x01, 0x42, 0xb5, 0x01,
	0x0a, 0x10, 0x63, 0x6f, 0x6d, 0x2e, 0x63, 0x68, 0x61, 0x70, 0x61, 0x72, 0x72, 0x61, 0x6c, 0x2e,
	0x76, 0x31, 0x42, 0x12, 0x4d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x3c, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x72, 0x65, 0x72, 0x69, 0x63, 0x6b, 0x73, 0x6f, 0x6e, 0x2f,
	0x63, 0x68, 0x61, 0x70, 0x61, 0x72, 0x72, 0x61, 0x6c, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x63, 0x68,
	0x61, 0x70, 0x61, 0x72, 0x72, 0x61, 0x6c, 0x2f, 0x76, 0x31, 0x3b, 0x63, 0x68, 0x61, 0x70, 0x61,
	0x72, 0x72, 0x61, 0x6c, 0x76, 0x31, 0xa2, 0x02, 0x03, 0x43, 0x58, 0x58, 0xaa, 0x02, 0x0c, 0x43,
	0x68, 0x61, 0x70, 0x61, 0x72, 0x72, 0x61, 0x6c, 0x2e, 0x56, 0x31, 0xca, 0x02, 0x0c, 0x43, 0x68,
	0x61, 0x70, 0x61, 0x72, 0x72, 0x61, 0x6c, 0x5c, 0x56, 0x31, 0xe2, 0x02, 0x18, 0x43, 0x68, 0x61,
	0x70, 0x61, 0x72, 0x72, 0x61, 0x6c, 0x5c, 0x56, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74,
	0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x0d, 0x43, 0x68, 0x61, 0x70, 0x61, 0x72, 0x72, 0x61,
	0x6c, 0x3a, 0x3a, 0x56, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_chaparral_v1_manage_service_proto_rawDescOnce sync.Once
	file_chaparral_v1_manage_service_proto_rawDescData = file_chaparral_v1_manage_service_proto_rawDesc
)

func file_chaparral_v1_manage_service_proto_rawDescGZIP() []byte {
	file_chaparral_v1_manage_service_proto_rawDescOnce.Do(func() {
		file_chaparral_v1_manage_service_proto_rawDescData = protoimpl.X.CompressGZIP(file_chaparral_v1_manage_service_proto_rawDescData)
	})
	return file_chaparral_v1_manage_service_proto_rawDescData
}

var file_chaparral_v1_manage_service_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_chaparral_v1_manage_service_proto_goTypes = []interface{}{
	(*StreamObjectRootsRequest)(nil),  // 0: chaparral.v1.StreamObjectRootsRequest
	(*StreamObjectRootsResponse)(nil), // 1: chaparral.v1.StreamObjectRootsResponse
}
var file_chaparral_v1_manage_service_proto_depIdxs = []int32{
	0, // 0: chaparral.v1.ManageService.StreamObjectRoots:input_type -> chaparral.v1.StreamObjectRootsRequest
	1, // 1: chaparral.v1.ManageService.StreamObjectRoots:output_type -> chaparral.v1.StreamObjectRootsResponse
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_chaparral_v1_manage_service_proto_init() }
func file_chaparral_v1_manage_service_proto_init() {
	if File_chaparral_v1_manage_service_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_chaparral_v1_manage_service_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StreamObjectRootsRequest); i {
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
		file_chaparral_v1_manage_service_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StreamObjectRootsResponse); i {
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
			RawDescriptor: file_chaparral_v1_manage_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_chaparral_v1_manage_service_proto_goTypes,
		DependencyIndexes: file_chaparral_v1_manage_service_proto_depIdxs,
		MessageInfos:      file_chaparral_v1_manage_service_proto_msgTypes,
	}.Build()
	File_chaparral_v1_manage_service_proto = out.File
	file_chaparral_v1_manage_service_proto_rawDesc = nil
	file_chaparral_v1_manage_service_proto_goTypes = nil
	file_chaparral_v1_manage_service_proto_depIdxs = nil
}