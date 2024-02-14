// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.32.0
// 	protoc        (unknown)
// source: chaparral/v1/access_service.proto

package chaparralv1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// GetObjectVersionRequest is used to request information about an object's state.
type GetObjectVersionRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The storage root id for the object to access. If not set, the default
	// storage root is used.
	StorageRootId string `protobuf:"bytes,1,opt,name=storage_root_id,json=storageRootId,proto3" json:"storage_root_id,omitempty"`
	// The object id to access (required).
	ObjectId string `protobuf:"bytes,2,opt,name=object_id,json=objectId,proto3" json:"object_id,omitempty"`
	// The version index for the object state. The default value is 0, which
	// refers to the most recent version.
	Version int32 `protobuf:"varint,3,opt,name=version,proto3" json:"version,omitempty"`
}

func (x *GetObjectVersionRequest) Reset() {
	*x = GetObjectVersionRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chaparral_v1_access_service_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetObjectVersionRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetObjectVersionRequest) ProtoMessage() {}

func (x *GetObjectVersionRequest) ProtoReflect() protoreflect.Message {
	mi := &file_chaparral_v1_access_service_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetObjectVersionRequest.ProtoReflect.Descriptor instead.
func (*GetObjectVersionRequest) Descriptor() ([]byte, []int) {
	return file_chaparral_v1_access_service_proto_rawDescGZIP(), []int{0}
}

func (x *GetObjectVersionRequest) GetStorageRootId() string {
	if x != nil {
		return x.StorageRootId
	}
	return ""
}

func (x *GetObjectVersionRequest) GetObjectId() string {
	if x != nil {
		return x.ObjectId
	}
	return ""
}

func (x *GetObjectVersionRequest) GetVersion() int32 {
	if x != nil {
		return x.Version
	}
	return 0
}

// GetObjectVersionResponse represents state for a specific object version.
type GetObjectVersionResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The object's storage root id. The empty string corresponds to the default
	// storage root.
	StorageRootId string `protobuf:"bytes,1,opt,name=storage_root_id,json=storageRootId,proto3" json:"storage_root_id,omitempty"`
	// The object's id
	ObjectId string `protobuf:"bytes,2,opt,name=object_id,json=objectId,proto3" json:"object_id,omitempty"`
	// The index for the object version represented by the state.
	Version int32 `protobuf:"varint,3,opt,name=version,proto3" json:"version,omitempty"`
	// The object's most recent version index.
	Head int32 `protobuf:"varint,4,opt,name=head,proto3" json:"head,omitempty"`
	// The object's digest algorithm (sha512 or sha256)
	DigestAlgorithm string `protobuf:"bytes,5,opt,name=digest_algorithm,json=digestAlgorithm,proto3" json:"digest_algorithm,omitempty"`
	// The object's logical state represented as a map from digests to
	// file info. Path entries in the file info represent logical
	// filenames for the object version state.
	State map[string]*FileInfo `protobuf:"bytes,6,rep,name=state,proto3" json:"state,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// The message associated with the object version
	Message string `protobuf:"bytes,7,opt,name=message,proto3" json:"message,omitempty"`
	// The user information associated with the object version
	User *User `protobuf:"bytes,8,opt,name=user,proto3" json:"user,omitempty"`
	// The timestamp associated witht he object version
	Created *timestamppb.Timestamp `protobuf:"bytes,9,opt,name=created,proto3" json:"created,omitempty"`
	// The OCFL specification version for the object version.
	Spec string `protobuf:"bytes,10,opt,name=spec,proto3" json:"spec,omitempty"`
}

func (x *GetObjectVersionResponse) Reset() {
	*x = GetObjectVersionResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chaparral_v1_access_service_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetObjectVersionResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetObjectVersionResponse) ProtoMessage() {}

func (x *GetObjectVersionResponse) ProtoReflect() protoreflect.Message {
	mi := &file_chaparral_v1_access_service_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetObjectVersionResponse.ProtoReflect.Descriptor instead.
func (*GetObjectVersionResponse) Descriptor() ([]byte, []int) {
	return file_chaparral_v1_access_service_proto_rawDescGZIP(), []int{1}
}

func (x *GetObjectVersionResponse) GetStorageRootId() string {
	if x != nil {
		return x.StorageRootId
	}
	return ""
}

func (x *GetObjectVersionResponse) GetObjectId() string {
	if x != nil {
		return x.ObjectId
	}
	return ""
}

func (x *GetObjectVersionResponse) GetVersion() int32 {
	if x != nil {
		return x.Version
	}
	return 0
}

func (x *GetObjectVersionResponse) GetHead() int32 {
	if x != nil {
		return x.Head
	}
	return 0
}

func (x *GetObjectVersionResponse) GetDigestAlgorithm() string {
	if x != nil {
		return x.DigestAlgorithm
	}
	return ""
}

func (x *GetObjectVersionResponse) GetState() map[string]*FileInfo {
	if x != nil {
		return x.State
	}
	return nil
}

func (x *GetObjectVersionResponse) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

func (x *GetObjectVersionResponse) GetUser() *User {
	if x != nil {
		return x.User
	}
	return nil
}

func (x *GetObjectVersionResponse) GetCreated() *timestamppb.Timestamp {
	if x != nil {
		return x.Created
	}
	return nil
}

func (x *GetObjectVersionResponse) GetSpec() string {
	if x != nil {
		return x.Spec
	}
	return ""
}

// GetObjectManifestRequest is used to request details about all content files
// in an object
type GetObjectManifestRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The storage root id for the object to access. If not set, the default
	// storage root is used.
	StorageRootId string `protobuf:"bytes,1,opt,name=storage_root_id,json=storageRootId,proto3" json:"storage_root_id,omitempty"`
	// The object id to access (required).
	ObjectId string `protobuf:"bytes,2,opt,name=object_id,json=objectId,proto3" json:"object_id,omitempty"`
}

func (x *GetObjectManifestRequest) Reset() {
	*x = GetObjectManifestRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chaparral_v1_access_service_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetObjectManifestRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetObjectManifestRequest) ProtoMessage() {}

func (x *GetObjectManifestRequest) ProtoReflect() protoreflect.Message {
	mi := &file_chaparral_v1_access_service_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetObjectManifestRequest.ProtoReflect.Descriptor instead.
func (*GetObjectManifestRequest) Descriptor() ([]byte, []int) {
	return file_chaparral_v1_access_service_proto_rawDescGZIP(), []int{2}
}

func (x *GetObjectManifestRequest) GetStorageRootId() string {
	if x != nil {
		return x.StorageRootId
	}
	return ""
}

func (x *GetObjectManifestRequest) GetObjectId() string {
	if x != nil {
		return x.ObjectId
	}
	return ""
}

// GetObjectManifestResponse represents all content files stored
// in an object across all versions
type GetObjectManifestResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The storage root id for the object
	StorageRootId string `protobuf:"bytes,1,opt,name=storage_root_id,json=storageRootId,proto3" json:"storage_root_id,omitempty"`
	// The object id for the manifest
	ObjectId string `protobuf:"bytes,2,opt,name=object_id,json=objectId,proto3" json:"object_id,omitempty"`
	// The object's path relative to the OCFL Storage Root
	// that contains it
	Path string `protobuf:"bytes,3,opt,name=path,proto3" json:"path,omitempty"`
	// digest algorithm used for manifest keys
	DigestAlgorithm string `protobuf:"bytes,4,opt,name=digest_algorithm,json=digestAlgorithm,proto3" json:"digest_algorithm,omitempty"`
	// manifest is a map of digest values to file info. Path
	// entries represent content paths relative to root of the
	// OCFL object
	Manifest map[string]*FileInfo `protobuf:"bytes,5,rep,name=manifest,proto3" json:"manifest,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// The OCFL specification version for the object
	Spec string `protobuf:"bytes,6,opt,name=spec,proto3" json:"spec,omitempty"`
}

func (x *GetObjectManifestResponse) Reset() {
	*x = GetObjectManifestResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chaparral_v1_access_service_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetObjectManifestResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetObjectManifestResponse) ProtoMessage() {}

func (x *GetObjectManifestResponse) ProtoReflect() protoreflect.Message {
	mi := &file_chaparral_v1_access_service_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetObjectManifestResponse.ProtoReflect.Descriptor instead.
func (*GetObjectManifestResponse) Descriptor() ([]byte, []int) {
	return file_chaparral_v1_access_service_proto_rawDescGZIP(), []int{3}
}

func (x *GetObjectManifestResponse) GetStorageRootId() string {
	if x != nil {
		return x.StorageRootId
	}
	return ""
}

func (x *GetObjectManifestResponse) GetObjectId() string {
	if x != nil {
		return x.ObjectId
	}
	return ""
}

func (x *GetObjectManifestResponse) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

func (x *GetObjectManifestResponse) GetDigestAlgorithm() string {
	if x != nil {
		return x.DigestAlgorithm
	}
	return ""
}

func (x *GetObjectManifestResponse) GetManifest() map[string]*FileInfo {
	if x != nil {
		return x.Manifest
	}
	return nil
}

func (x *GetObjectManifestResponse) GetSpec() string {
	if x != nil {
		return x.Spec
	}
	return ""
}

type FileInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// file size
	Size int64 `protobuf:"varint,1,opt,name=size,proto3" json:"size,omitempty"`
	// one or more file paths for the content
	Paths []string `protobuf:"bytes,2,rep,name=paths,proto3" json:"paths,omitempty"`
	// map of alternate digests alg -> digest
	Fixity map[string]string `protobuf:"bytes,3,rep,name=fixity,proto3" json:"fixity,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *FileInfo) Reset() {
	*x = FileInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_chaparral_v1_access_service_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FileInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FileInfo) ProtoMessage() {}

func (x *FileInfo) ProtoReflect() protoreflect.Message {
	mi := &file_chaparral_v1_access_service_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FileInfo.ProtoReflect.Descriptor instead.
func (*FileInfo) Descriptor() ([]byte, []int) {
	return file_chaparral_v1_access_service_proto_rawDescGZIP(), []int{4}
}

func (x *FileInfo) GetSize() int64 {
	if x != nil {
		return x.Size
	}
	return 0
}

func (x *FileInfo) GetPaths() []string {
	if x != nil {
		return x.Paths
	}
	return nil
}

func (x *FileInfo) GetFixity() map[string]string {
	if x != nil {
		return x.Fixity
	}
	return nil
}

var File_chaparral_v1_access_service_proto protoreflect.FileDescriptor

var file_chaparral_v1_access_service_proto_rawDesc = []byte{
	0x0a, 0x21, 0x63, 0x68, 0x61, 0x70, 0x61, 0x72, 0x72, 0x61, 0x6c, 0x2f, 0x76, 0x31, 0x2f, 0x61,
	0x63, 0x63, 0x65, 0x73, 0x73, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x0c, 0x63, 0x68, 0x61, 0x70, 0x61, 0x72, 0x72, 0x61, 0x6c, 0x2e, 0x76,
	0x31, 0x1a, 0x17, 0x63, 0x68, 0x61, 0x70, 0x61, 0x72, 0x72, 0x61, 0x6c, 0x2f, 0x76, 0x31, 0x2f,
	0x63, 0x6f, 0x72, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65,
	0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x78, 0x0a, 0x17, 0x47,
	0x65, 0x74, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x26, 0x0a, 0x0f, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67,
	0x65, 0x5f, 0x72, 0x6f, 0x6f, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0d, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x52, 0x6f, 0x6f, 0x74, 0x49, 0x64, 0x12, 0x1b,
	0x0a, 0x09, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x08, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x49, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x76,
	0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x07, 0x76, 0x65,
	0x72, 0x73, 0x69, 0x6f, 0x6e, 0x22, 0xdf, 0x03, 0x0a, 0x18, 0x47, 0x65, 0x74, 0x4f, 0x62, 0x6a,
	0x65, 0x63, 0x74, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x26, 0x0a, 0x0f, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x5f, 0x72, 0x6f,
	0x6f, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x73, 0x74, 0x6f,
	0x72, 0x61, 0x67, 0x65, 0x52, 0x6f, 0x6f, 0x74, 0x49, 0x64, 0x12, 0x1b, 0x0a, 0x09, 0x6f, 0x62,
	0x6a, 0x65, 0x63, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x6f,
	0x62, 0x6a, 0x65, 0x63, 0x74, 0x49, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69,
	0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f,
	0x6e, 0x12, 0x12, 0x0a, 0x04, 0x68, 0x65, 0x61, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52,
	0x04, 0x68, 0x65, 0x61, 0x64, 0x12, 0x29, 0x0a, 0x10, 0x64, 0x69, 0x67, 0x65, 0x73, 0x74, 0x5f,
	0x61, 0x6c, 0x67, 0x6f, 0x72, 0x69, 0x74, 0x68, 0x6d, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0f, 0x64, 0x69, 0x67, 0x65, 0x73, 0x74, 0x41, 0x6c, 0x67, 0x6f, 0x72, 0x69, 0x74, 0x68, 0x6d,
	0x12, 0x47, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x18, 0x06, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x31, 0x2e, 0x63, 0x68, 0x61, 0x70, 0x61, 0x72, 0x72, 0x61, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x47,
	0x65, 0x74, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x65, 0x45, 0x6e, 0x74,
	0x72, 0x79, 0x52, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x12, 0x26, 0x0a, 0x04, 0x75, 0x73, 0x65, 0x72, 0x18, 0x08, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x12, 0x2e, 0x63, 0x68, 0x61, 0x70, 0x61, 0x72, 0x72, 0x61, 0x6c, 0x2e, 0x76, 0x31,
	0x2e, 0x55, 0x73, 0x65, 0x72, 0x52, 0x04, 0x75, 0x73, 0x65, 0x72, 0x12, 0x34, 0x0a, 0x07, 0x63,
	0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x18, 0x09, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54,
	0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x07, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65,
	0x64, 0x12, 0x12, 0x0a, 0x04, 0x73, 0x70, 0x65, 0x63, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x73, 0x70, 0x65, 0x63, 0x1a, 0x50, 0x0a, 0x0a, 0x53, 0x74, 0x61, 0x74, 0x65, 0x45, 0x6e,
	0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x2c, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x63, 0x68, 0x61, 0x70, 0x61, 0x72, 0x72, 0x61, 0x6c,
	0x2e, 0x76, 0x31, 0x2e, 0x46, 0x69, 0x6c, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x05, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x5f, 0x0a, 0x18, 0x47, 0x65, 0x74, 0x4f, 0x62,
	0x6a, 0x65, 0x63, 0x74, 0x4d, 0x61, 0x6e, 0x69, 0x66, 0x65, 0x73, 0x74, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x26, 0x0a, 0x0f, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x5f, 0x72,
	0x6f, 0x6f, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x73, 0x74,
	0x6f, 0x72, 0x61, 0x67, 0x65, 0x52, 0x6f, 0x6f, 0x74, 0x49, 0x64, 0x12, 0x1b, 0x0a, 0x09, 0x6f,
	0x62, 0x6a, 0x65, 0x63, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08,
	0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x49, 0x64, 0x22, 0xdb, 0x02, 0x0a, 0x19, 0x47, 0x65, 0x74,
	0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x4d, 0x61, 0x6e, 0x69, 0x66, 0x65, 0x73, 0x74, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x26, 0x0a, 0x0f, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67,
	0x65, 0x5f, 0x72, 0x6f, 0x6f, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0d, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x52, 0x6f, 0x6f, 0x74, 0x49, 0x64, 0x12, 0x1b,
	0x0a, 0x09, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x08, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x49, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x70,
	0x61, 0x74, 0x68, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x70, 0x61, 0x74, 0x68, 0x12,
	0x29, 0x0a, 0x10, 0x64, 0x69, 0x67, 0x65, 0x73, 0x74, 0x5f, 0x61, 0x6c, 0x67, 0x6f, 0x72, 0x69,
	0x74, 0x68, 0x6d, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0f, 0x64, 0x69, 0x67, 0x65, 0x73,
	0x74, 0x41, 0x6c, 0x67, 0x6f, 0x72, 0x69, 0x74, 0x68, 0x6d, 0x12, 0x51, 0x0a, 0x08, 0x6d, 0x61,
	0x6e, 0x69, 0x66, 0x65, 0x73, 0x74, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x35, 0x2e, 0x63,
	0x68, 0x61, 0x70, 0x61, 0x72, 0x72, 0x61, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x4f,
	0x62, 0x6a, 0x65, 0x63, 0x74, 0x4d, 0x61, 0x6e, 0x69, 0x66, 0x65, 0x73, 0x74, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x4d, 0x61, 0x6e, 0x69, 0x66, 0x65, 0x73, 0x74, 0x45, 0x6e,
	0x74, 0x72, 0x79, 0x52, 0x08, 0x6d, 0x61, 0x6e, 0x69, 0x66, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a,
	0x04, 0x73, 0x70, 0x65, 0x63, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x73, 0x70, 0x65,
	0x63, 0x1a, 0x53, 0x0a, 0x0d, 0x4d, 0x61, 0x6e, 0x69, 0x66, 0x65, 0x73, 0x74, 0x45, 0x6e, 0x74,
	0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x03, 0x6b, 0x65, 0x79, 0x12, 0x2c, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x63, 0x68, 0x61, 0x70, 0x61, 0x72, 0x72, 0x61, 0x6c, 0x2e,
	0x76, 0x31, 0x2e, 0x46, 0x69, 0x6c, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x05, 0x76, 0x61, 0x6c,
	0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0xab, 0x01, 0x0a, 0x08, 0x46, 0x69, 0x6c, 0x65, 0x49,
	0x6e, 0x66, 0x6f, 0x12, 0x12, 0x0a, 0x04, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x03, 0x52, 0x04, 0x73, 0x69, 0x7a, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x70, 0x61, 0x74, 0x68, 0x73,
	0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x52, 0x05, 0x70, 0x61, 0x74, 0x68, 0x73, 0x12, 0x3a, 0x0a,
	0x06, 0x66, 0x69, 0x78, 0x69, 0x74, 0x79, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x22, 0x2e,
	0x63, 0x68, 0x61, 0x70, 0x61, 0x72, 0x72, 0x61, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x46, 0x69, 0x6c,
	0x65, 0x49, 0x6e, 0x66, 0x6f, 0x2e, 0x46, 0x69, 0x78, 0x69, 0x74, 0x79, 0x45, 0x6e, 0x74, 0x72,
	0x79, 0x52, 0x06, 0x66, 0x69, 0x78, 0x69, 0x74, 0x79, 0x1a, 0x39, 0x0a, 0x0b, 0x46, 0x69, 0x78,
	0x69, 0x74, 0x79, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x3a, 0x02, 0x38, 0x01, 0x32, 0xdc, 0x01, 0x0a, 0x0d, 0x41, 0x63, 0x63, 0x65, 0x73, 0x73, 0x53,
	0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x63, 0x0a, 0x10, 0x47, 0x65, 0x74, 0x4f, 0x62, 0x6a,
	0x65, 0x63, 0x74, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x25, 0x2e, 0x63, 0x68, 0x61,
	0x70, 0x61, 0x72, 0x72, 0x61, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x4f, 0x62, 0x6a,
	0x65, 0x63, 0x74, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x26, 0x2e, 0x63, 0x68, 0x61, 0x70, 0x61, 0x72, 0x72, 0x61, 0x6c, 0x2e, 0x76, 0x31,
	0x2e, 0x47, 0x65, 0x74, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f,
	0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x66, 0x0a, 0x11, 0x47,
	0x65, 0x74, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x4d, 0x61, 0x6e, 0x69, 0x66, 0x65, 0x73, 0x74,
	0x12, 0x26, 0x2e, 0x63, 0x68, 0x61, 0x70, 0x61, 0x72, 0x72, 0x61, 0x6c, 0x2e, 0x76, 0x31, 0x2e,
	0x47, 0x65, 0x74, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x4d, 0x61, 0x6e, 0x69, 0x66, 0x65, 0x73,
	0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x27, 0x2e, 0x63, 0x68, 0x61, 0x70, 0x61,
	0x72, 0x72, 0x61, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x4f, 0x62, 0x6a, 0x65, 0x63,
	0x74, 0x4d, 0x61, 0x6e, 0x69, 0x66, 0x65, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x22, 0x00, 0x42, 0xb5, 0x01, 0x0a, 0x10, 0x63, 0x6f, 0x6d, 0x2e, 0x63, 0x68, 0x61, 0x70,
	0x61, 0x72, 0x72, 0x61, 0x6c, 0x2e, 0x76, 0x31, 0x42, 0x12, 0x41, 0x63, 0x63, 0x65, 0x73, 0x73,
	0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x3c,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x72, 0x65, 0x72, 0x69,
	0x63, 0x6b, 0x73, 0x6f, 0x6e, 0x2f, 0x63, 0x68, 0x61, 0x70, 0x61, 0x72, 0x72, 0x61, 0x6c, 0x2f,
	0x67, 0x65, 0x6e, 0x2f, 0x63, 0x68, 0x61, 0x70, 0x61, 0x72, 0x72, 0x61, 0x6c, 0x2f, 0x76, 0x31,
	0x3b, 0x63, 0x68, 0x61, 0x70, 0x61, 0x72, 0x72, 0x61, 0x6c, 0x76, 0x31, 0xa2, 0x02, 0x03, 0x43,
	0x58, 0x58, 0xaa, 0x02, 0x0c, 0x43, 0x68, 0x61, 0x70, 0x61, 0x72, 0x72, 0x61, 0x6c, 0x2e, 0x56,
	0x31, 0xca, 0x02, 0x0c, 0x43, 0x68, 0x61, 0x70, 0x61, 0x72, 0x72, 0x61, 0x6c, 0x5c, 0x56, 0x31,
	0xe2, 0x02, 0x18, 0x43, 0x68, 0x61, 0x70, 0x61, 0x72, 0x72, 0x61, 0x6c, 0x5c, 0x56, 0x31, 0x5c,
	0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x0d, 0x43, 0x68,
	0x61, 0x70, 0x61, 0x72, 0x72, 0x61, 0x6c, 0x3a, 0x3a, 0x56, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_chaparral_v1_access_service_proto_rawDescOnce sync.Once
	file_chaparral_v1_access_service_proto_rawDescData = file_chaparral_v1_access_service_proto_rawDesc
)

func file_chaparral_v1_access_service_proto_rawDescGZIP() []byte {
	file_chaparral_v1_access_service_proto_rawDescOnce.Do(func() {
		file_chaparral_v1_access_service_proto_rawDescData = protoimpl.X.CompressGZIP(file_chaparral_v1_access_service_proto_rawDescData)
	})
	return file_chaparral_v1_access_service_proto_rawDescData
}

var file_chaparral_v1_access_service_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_chaparral_v1_access_service_proto_goTypes = []interface{}{
	(*GetObjectVersionRequest)(nil),   // 0: chaparral.v1.GetObjectVersionRequest
	(*GetObjectVersionResponse)(nil),  // 1: chaparral.v1.GetObjectVersionResponse
	(*GetObjectManifestRequest)(nil),  // 2: chaparral.v1.GetObjectManifestRequest
	(*GetObjectManifestResponse)(nil), // 3: chaparral.v1.GetObjectManifestResponse
	(*FileInfo)(nil),                  // 4: chaparral.v1.FileInfo
	nil,                               // 5: chaparral.v1.GetObjectVersionResponse.StateEntry
	nil,                               // 6: chaparral.v1.GetObjectManifestResponse.ManifestEntry
	nil,                               // 7: chaparral.v1.FileInfo.FixityEntry
	(*User)(nil),                      // 8: chaparral.v1.User
	(*timestamppb.Timestamp)(nil),     // 9: google.protobuf.Timestamp
}
var file_chaparral_v1_access_service_proto_depIdxs = []int32{
	5, // 0: chaparral.v1.GetObjectVersionResponse.state:type_name -> chaparral.v1.GetObjectVersionResponse.StateEntry
	8, // 1: chaparral.v1.GetObjectVersionResponse.user:type_name -> chaparral.v1.User
	9, // 2: chaparral.v1.GetObjectVersionResponse.created:type_name -> google.protobuf.Timestamp
	6, // 3: chaparral.v1.GetObjectManifestResponse.manifest:type_name -> chaparral.v1.GetObjectManifestResponse.ManifestEntry
	7, // 4: chaparral.v1.FileInfo.fixity:type_name -> chaparral.v1.FileInfo.FixityEntry
	4, // 5: chaparral.v1.GetObjectVersionResponse.StateEntry.value:type_name -> chaparral.v1.FileInfo
	4, // 6: chaparral.v1.GetObjectManifestResponse.ManifestEntry.value:type_name -> chaparral.v1.FileInfo
	0, // 7: chaparral.v1.AccessService.GetObjectVersion:input_type -> chaparral.v1.GetObjectVersionRequest
	2, // 8: chaparral.v1.AccessService.GetObjectManifest:input_type -> chaparral.v1.GetObjectManifestRequest
	1, // 9: chaparral.v1.AccessService.GetObjectVersion:output_type -> chaparral.v1.GetObjectVersionResponse
	3, // 10: chaparral.v1.AccessService.GetObjectManifest:output_type -> chaparral.v1.GetObjectManifestResponse
	9, // [9:11] is the sub-list for method output_type
	7, // [7:9] is the sub-list for method input_type
	7, // [7:7] is the sub-list for extension type_name
	7, // [7:7] is the sub-list for extension extendee
	0, // [0:7] is the sub-list for field type_name
}

func init() { file_chaparral_v1_access_service_proto_init() }
func file_chaparral_v1_access_service_proto_init() {
	if File_chaparral_v1_access_service_proto != nil {
		return
	}
	file_chaparral_v1_core_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_chaparral_v1_access_service_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetObjectVersionRequest); i {
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
		file_chaparral_v1_access_service_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetObjectVersionResponse); i {
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
		file_chaparral_v1_access_service_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetObjectManifestRequest); i {
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
		file_chaparral_v1_access_service_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetObjectManifestResponse); i {
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
		file_chaparral_v1_access_service_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FileInfo); i {
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
			RawDescriptor: file_chaparral_v1_access_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_chaparral_v1_access_service_proto_goTypes,
		DependencyIndexes: file_chaparral_v1_access_service_proto_depIdxs,
		MessageInfos:      file_chaparral_v1_access_service_proto_msgTypes,
	}.Build()
	File_chaparral_v1_access_service_proto = out.File
	file_chaparral_v1_access_service_proto_rawDesc = nil
	file_chaparral_v1_access_service_proto_goTypes = nil
	file_chaparral_v1_access_service_proto_depIdxs = nil
}
