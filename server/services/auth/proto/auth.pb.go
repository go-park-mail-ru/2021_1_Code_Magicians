// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.6.1
// source: auth.proto

package __

import (
	context "context"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
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

type UserAuth struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Username string `protobuf:"bytes,1,opt,name=Username,proto3" json:"Username,omitempty"`
	Password string `protobuf:"bytes,2,opt,name=Password,proto3" json:"Password,omitempty"`
}

func (x *UserAuth) Reset() {
	*x = UserAuth{}
	if protoimpl.UnsafeEnabled {
		mi := &file_auth_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UserAuth) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UserAuth) ProtoMessage() {}

func (x *UserAuth) ProtoReflect() protoreflect.Message {
	mi := &file_auth_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UserAuth.ProtoReflect.Descriptor instead.
func (*UserAuth) Descriptor() ([]byte, []int) {
	return file_auth_proto_rawDescGZIP(), []int{0}
}

func (x *UserAuth) GetUsername() string {
	if x != nil {
		return x.Username
	}
	return ""
}

func (x *UserAuth) GetPassword() string {
	if x != nil {
		return x.Password
	}
	return ""
}

type VkToken struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Token string `protobuf:"bytes,1,opt,name=Token,proto3" json:"Token,omitempty"`
}

func (x *VkToken) Reset() {
	*x = VkToken{}
	if protoimpl.UnsafeEnabled {
		mi := &file_auth_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *VkToken) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*VkToken) ProtoMessage() {}

func (x *VkToken) ProtoReflect() protoreflect.Message {
	mi := &file_auth_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use VkToken.ProtoReflect.Descriptor instead.
func (*VkToken) Descriptor() ([]byte, []int) {
	return file_auth_proto_rawDescGZIP(), []int{1}
}

func (x *VkToken) GetToken() string {
	if x != nil {
		return x.Token
	}
	return ""
}

type VkTokenInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	UserID  int64                `protobuf:"varint,1,opt,name=UserID,proto3" json:"UserID,omitempty"`
	Token   string               `protobuf:"bytes,2,opt,name=Token,proto3" json:"Token,omitempty"`
	Expires *timestamp.Timestamp `protobuf:"bytes,3,opt,name=Expires,proto3" json:"Expires,omitempty"`
}

func (x *VkTokenInfo) Reset() {
	*x = VkTokenInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_auth_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *VkTokenInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*VkTokenInfo) ProtoMessage() {}

func (x *VkTokenInfo) ProtoReflect() protoreflect.Message {
	mi := &file_auth_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use VkTokenInfo.ProtoReflect.Descriptor instead.
func (*VkTokenInfo) Descriptor() ([]byte, []int) {
	return file_auth_proto_rawDescGZIP(), []int{2}
}

func (x *VkTokenInfo) GetUserID() int64 {
	if x != nil {
		return x.UserID
	}
	return 0
}

func (x *VkTokenInfo) GetToken() string {
	if x != nil {
		return x.Token
	}
	return ""
}

func (x *VkTokenInfo) GetExpires() *timestamp.Timestamp {
	if x != nil {
		return x.Expires
	}
	return nil
}

type CookieValue struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	CookieValue string `protobuf:"bytes,1,opt,name=cookieValue,proto3" json:"cookieValue,omitempty"`
}

func (x *CookieValue) Reset() {
	*x = CookieValue{}
	if protoimpl.UnsafeEnabled {
		mi := &file_auth_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CookieValue) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CookieValue) ProtoMessage() {}

func (x *CookieValue) ProtoReflect() protoreflect.Message {
	mi := &file_auth_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CookieValue.ProtoReflect.Descriptor instead.
func (*CookieValue) Descriptor() ([]byte, []int) {
	return file_auth_proto_rawDescGZIP(), []int{3}
}

func (x *CookieValue) GetCookieValue() string {
	if x != nil {
		return x.CookieValue
	}
	return ""
}

type UserID struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Uid int64 `protobuf:"varint,1,opt,name=uid,proto3" json:"uid,omitempty"`
}

func (x *UserID) Reset() {
	*x = UserID{}
	if protoimpl.UnsafeEnabled {
		mi := &file_auth_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UserID) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UserID) ProtoMessage() {}

func (x *UserID) ProtoReflect() protoreflect.Message {
	mi := &file_auth_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UserID.ProtoReflect.Descriptor instead.
func (*UserID) Descriptor() ([]byte, []int) {
	return file_auth_proto_rawDescGZIP(), []int{4}
}

func (x *UserID) GetUid() int64 {
	if x != nil {
		return x.Uid
	}
	return 0
}

type Cookie struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Value   string               `protobuf:"bytes,1,opt,name=Value,proto3" json:"Value,omitempty"`
	Expires *timestamp.Timestamp `protobuf:"bytes,2,opt,name=Expires,proto3" json:"Expires,omitempty"`
}

func (x *Cookie) Reset() {
	*x = Cookie{}
	if protoimpl.UnsafeEnabled {
		mi := &file_auth_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Cookie) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Cookie) ProtoMessage() {}

func (x *Cookie) ProtoReflect() protoreflect.Message {
	mi := &file_auth_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Cookie.ProtoReflect.Descriptor instead.
func (*Cookie) Descriptor() ([]byte, []int) {
	return file_auth_proto_rawDescGZIP(), []int{5}
}

func (x *Cookie) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

func (x *Cookie) GetExpires() *timestamp.Timestamp {
	if x != nil {
		return x.Expires
	}
	return nil
}

type CookieInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	UserID int64   `protobuf:"varint,1,opt,name=userID,proto3" json:"userID,omitempty"`
	Cookie *Cookie `protobuf:"bytes,2,opt,name=cookie,proto3" json:"cookie,omitempty"`
}

func (x *CookieInfo) Reset() {
	*x = CookieInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_auth_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CookieInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CookieInfo) ProtoMessage() {}

func (x *CookieInfo) ProtoReflect() protoreflect.Message {
	mi := &file_auth_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CookieInfo.ProtoReflect.Descriptor instead.
func (*CookieInfo) Descriptor() ([]byte, []int) {
	return file_auth_proto_rawDescGZIP(), []int{6}
}

func (x *CookieInfo) GetUserID() int64 {
	if x != nil {
		return x.UserID
	}
	return 0
}

func (x *CookieInfo) GetCookie() *Cookie {
	if x != nil {
		return x.Cookie
	}
	return nil
}

type Error struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Error) Reset() {
	*x = Error{}
	if protoimpl.UnsafeEnabled {
		mi := &file_auth_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Error) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Error) ProtoMessage() {}

func (x *Error) ProtoReflect() protoreflect.Message {
	mi := &file_auth_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Error.ProtoReflect.Descriptor instead.
func (*Error) Descriptor() ([]byte, []int) {
	return file_auth_proto_rawDescGZIP(), []int{7}
}

var File_auth_proto protoreflect.FileDescriptor

var file_auth_proto_rawDesc = []byte{
	0x0a, 0x0a, 0x61, 0x75, 0x74, 0x68, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x04, 0x61, 0x75,
	0x74, 0x68, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0x42, 0x0a, 0x08, 0x55, 0x73, 0x65, 0x72, 0x41, 0x75, 0x74, 0x68, 0x12,
	0x1a, 0x0a, 0x08, 0x55, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x08, 0x55, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x50,
	0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x50,
	0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x22, 0x1f, 0x0a, 0x07, 0x56, 0x6b, 0x54, 0x6f, 0x6b,
	0x65, 0x6e, 0x12, 0x14, 0x0a, 0x05, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x05, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x22, 0x71, 0x0a, 0x0b, 0x56, 0x6b, 0x54, 0x6f,
	0x6b, 0x65, 0x6e, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x16, 0x0a, 0x06, 0x55, 0x73, 0x65, 0x72, 0x49,
	0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x55, 0x73, 0x65, 0x72, 0x49, 0x44, 0x12,
	0x14, 0x0a, 0x05, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05,
	0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x12, 0x34, 0x0a, 0x07, 0x45, 0x78, 0x70, 0x69, 0x72, 0x65, 0x73,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61,
	0x6d, 0x70, 0x52, 0x07, 0x45, 0x78, 0x70, 0x69, 0x72, 0x65, 0x73, 0x22, 0x2f, 0x0a, 0x0b, 0x43,
	0x6f, 0x6f, 0x6b, 0x69, 0x65, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x20, 0x0a, 0x0b, 0x63, 0x6f,
	0x6f, 0x6b, 0x69, 0x65, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0b, 0x63, 0x6f, 0x6f, 0x6b, 0x69, 0x65, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x1a, 0x0a, 0x06,
	0x55, 0x73, 0x65, 0x72, 0x49, 0x44, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x03, 0x52, 0x03, 0x75, 0x69, 0x64, 0x22, 0x54, 0x0a, 0x06, 0x43, 0x6f, 0x6f, 0x6b,
	0x69, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x05, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x34, 0x0a, 0x07, 0x45, 0x78, 0x70, 0x69,
	0x72, 0x65, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65,
	0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x07, 0x45, 0x78, 0x70, 0x69, 0x72, 0x65, 0x73, 0x22, 0x4a,
	0x0a, 0x0a, 0x43, 0x6f, 0x6f, 0x6b, 0x69, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x16, 0x0a, 0x06,
	0x75, 0x73, 0x65, 0x72, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x75, 0x73,
	0x65, 0x72, 0x49, 0x44, 0x12, 0x24, 0x0a, 0x06, 0x63, 0x6f, 0x6f, 0x6b, 0x69, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x0c, 0x2e, 0x61, 0x75, 0x74, 0x68, 0x2e, 0x43, 0x6f, 0x6f, 0x6b,
	0x69, 0x65, 0x52, 0x06, 0x63, 0x6f, 0x6f, 0x6b, 0x69, 0x65, 0x22, 0x07, 0x0a, 0x05, 0x45, 0x72,
	0x72, 0x6f, 0x72, 0x32, 0xa4, 0x03, 0x0a, 0x04, 0x41, 0x75, 0x74, 0x68, 0x12, 0x35, 0x0a, 0x14,
	0x43, 0x68, 0x65, 0x63, 0x6b, 0x55, 0x73, 0x65, 0x72, 0x43, 0x72, 0x65, 0x64, 0x65, 0x6e, 0x74,
	0x69, 0x61, 0x6c, 0x73, 0x12, 0x0e, 0x2e, 0x61, 0x75, 0x74, 0x68, 0x2e, 0x55, 0x73, 0x65, 0x72,
	0x41, 0x75, 0x74, 0x68, 0x1a, 0x0b, 0x2e, 0x61, 0x75, 0x74, 0x68, 0x2e, 0x45, 0x72, 0x72, 0x6f,
	0x72, 0x22, 0x00, 0x12, 0x30, 0x0a, 0x0d, 0x41, 0x64, 0x64, 0x43, 0x6f, 0x6f, 0x6b, 0x69, 0x65,
	0x49, 0x6e, 0x66, 0x6f, 0x12, 0x10, 0x2e, 0x61, 0x75, 0x74, 0x68, 0x2e, 0x43, 0x6f, 0x6f, 0x6b,
	0x69, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x1a, 0x0b, 0x2e, 0x61, 0x75, 0x74, 0x68, 0x2e, 0x45, 0x72,
	0x72, 0x6f, 0x72, 0x22, 0x00, 0x12, 0x36, 0x0a, 0x0d, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x42,
	0x79, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x11, 0x2e, 0x61, 0x75, 0x74, 0x68, 0x2e, 0x43, 0x6f,
	0x6f, 0x6b, 0x69, 0x65, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x1a, 0x10, 0x2e, 0x61, 0x75, 0x74, 0x68,
	0x2e, 0x43, 0x6f, 0x6f, 0x6b, 0x69, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x22, 0x00, 0x12, 0x32, 0x0a,
	0x0e, 0x53, 0x65, 0x61, 0x72, 0x63, 0x68, 0x42, 0x79, 0x55, 0x73, 0x65, 0x72, 0x49, 0x44, 0x12,
	0x0c, 0x2e, 0x61, 0x75, 0x74, 0x68, 0x2e, 0x55, 0x73, 0x65, 0x72, 0x49, 0x44, 0x1a, 0x10, 0x2e,
	0x61, 0x75, 0x74, 0x68, 0x2e, 0x43, 0x6f, 0x6f, 0x6b, 0x69, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x22,
	0x00, 0x12, 0x2f, 0x0a, 0x0c, 0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x43, 0x6f, 0x6f, 0x6b, 0x69,
	0x65, 0x12, 0x10, 0x2e, 0x61, 0x75, 0x74, 0x68, 0x2e, 0x43, 0x6f, 0x6f, 0x6b, 0x69, 0x65, 0x49,
	0x6e, 0x66, 0x6f, 0x1a, 0x0b, 0x2e, 0x61, 0x75, 0x74, 0x68, 0x2e, 0x45, 0x72, 0x72, 0x6f, 0x72,
	0x22, 0x00, 0x12, 0x33, 0x0a, 0x12, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x55, 0x73, 0x65, 0x72, 0x42,
	0x79, 0x56, 0x6b, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x12, 0x0d, 0x2e, 0x61, 0x75, 0x74, 0x68, 0x2e,
	0x56, 0x6b, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x1a, 0x0c, 0x2e, 0x61, 0x75, 0x74, 0x68, 0x2e, 0x55,
	0x73, 0x65, 0x72, 0x49, 0x44, 0x22, 0x00, 0x12, 0x2e, 0x0a, 0x0a, 0x41, 0x64, 0x64, 0x56, 0x6b,
	0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x12, 0x11, 0x2e, 0x61, 0x75, 0x74, 0x68, 0x2e, 0x56, 0x6b, 0x54,
	0x6f, 0x6b, 0x65, 0x6e, 0x49, 0x6e, 0x66, 0x6f, 0x1a, 0x0b, 0x2e, 0x61, 0x75, 0x74, 0x68, 0x2e,
	0x45, 0x72, 0x72, 0x6f, 0x72, 0x22, 0x00, 0x12, 0x31, 0x0a, 0x0d, 0x52, 0x65, 0x6d, 0x6f, 0x76,
	0x65, 0x56, 0x6b, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x12, 0x11, 0x2e, 0x61, 0x75, 0x74, 0x68, 0x2e,
	0x56, 0x6b, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x49, 0x6e, 0x66, 0x6f, 0x1a, 0x0b, 0x2e, 0x61, 0x75,
	0x74, 0x68, 0x2e, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x22, 0x00, 0x42, 0x04, 0x5a, 0x02, 0x2e, 0x2f,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_auth_proto_rawDescOnce sync.Once
	file_auth_proto_rawDescData = file_auth_proto_rawDesc
)

func file_auth_proto_rawDescGZIP() []byte {
	file_auth_proto_rawDescOnce.Do(func() {
		file_auth_proto_rawDescData = protoimpl.X.CompressGZIP(file_auth_proto_rawDescData)
	})
	return file_auth_proto_rawDescData
}

var file_auth_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_auth_proto_goTypes = []interface{}{
	(*UserAuth)(nil),            // 0: auth.UserAuth
	(*VkToken)(nil),             // 1: auth.VkToken
	(*VkTokenInfo)(nil),         // 2: auth.VkTokenInfo
	(*CookieValue)(nil),         // 3: auth.CookieValue
	(*UserID)(nil),              // 4: auth.UserID
	(*Cookie)(nil),              // 5: auth.Cookie
	(*CookieInfo)(nil),          // 6: auth.CookieInfo
	(*Error)(nil),               // 7: auth.Error
	(*timestamp.Timestamp)(nil), // 8: google.protobuf.Timestamp
}
var file_auth_proto_depIdxs = []int32{
	8,  // 0: auth.VkTokenInfo.Expires:type_name -> google.protobuf.Timestamp
	8,  // 1: auth.Cookie.Expires:type_name -> google.protobuf.Timestamp
	5,  // 2: auth.CookieInfo.cookie:type_name -> auth.Cookie
	0,  // 3: auth.Auth.CheckUserCredentials:input_type -> auth.UserAuth
	6,  // 4: auth.Auth.AddCookieInfo:input_type -> auth.CookieInfo
	3,  // 5: auth.Auth.SearchByValue:input_type -> auth.CookieValue
	4,  // 6: auth.Auth.SearchByUserID:input_type -> auth.UserID
	6,  // 7: auth.Auth.RemoveCookie:input_type -> auth.CookieInfo
	1,  // 8: auth.Auth.CheckUserByVkToken:input_type -> auth.VkToken
	2,  // 9: auth.Auth.AddVkToken:input_type -> auth.VkTokenInfo
	2,  // 10: auth.Auth.RemoveVkToken:input_type -> auth.VkTokenInfo
	7,  // 11: auth.Auth.CheckUserCredentials:output_type -> auth.Error
	7,  // 12: auth.Auth.AddCookieInfo:output_type -> auth.Error
	6,  // 13: auth.Auth.SearchByValue:output_type -> auth.CookieInfo
	6,  // 14: auth.Auth.SearchByUserID:output_type -> auth.CookieInfo
	7,  // 15: auth.Auth.RemoveCookie:output_type -> auth.Error
	4,  // 16: auth.Auth.CheckUserByVkToken:output_type -> auth.UserID
	7,  // 17: auth.Auth.AddVkToken:output_type -> auth.Error
	7,  // 18: auth.Auth.RemoveVkToken:output_type -> auth.Error
	11, // [11:19] is the sub-list for method output_type
	3,  // [3:11] is the sub-list for method input_type
	3,  // [3:3] is the sub-list for extension type_name
	3,  // [3:3] is the sub-list for extension extendee
	0,  // [0:3] is the sub-list for field type_name
}

func init() { file_auth_proto_init() }
func file_auth_proto_init() {
	if File_auth_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_auth_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UserAuth); i {
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
		file_auth_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*VkToken); i {
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
		file_auth_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*VkTokenInfo); i {
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
		file_auth_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CookieValue); i {
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
		file_auth_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UserID); i {
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
		file_auth_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Cookie); i {
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
		file_auth_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CookieInfo); i {
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
		file_auth_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Error); i {
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
			RawDescriptor: file_auth_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_auth_proto_goTypes,
		DependencyIndexes: file_auth_proto_depIdxs,
		MessageInfos:      file_auth_proto_msgTypes,
	}.Build()
	File_auth_proto = out.File
	file_auth_proto_rawDesc = nil
	file_auth_proto_goTypes = nil
	file_auth_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// AuthClient is the client API for Auth service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type AuthClient interface {
	CheckUserCredentials(ctx context.Context, in *UserAuth, opts ...grpc.CallOption) (*Error, error)
	AddCookieInfo(ctx context.Context, in *CookieInfo, opts ...grpc.CallOption) (*Error, error)
	SearchByValue(ctx context.Context, in *CookieValue, opts ...grpc.CallOption) (*CookieInfo, error)
	SearchByUserID(ctx context.Context, in *UserID, opts ...grpc.CallOption) (*CookieInfo, error)
	RemoveCookie(ctx context.Context, in *CookieInfo, opts ...grpc.CallOption) (*Error, error)
	CheckUserByVkToken(ctx context.Context, in *VkToken, opts ...grpc.CallOption) (*UserID, error)
	AddVkToken(ctx context.Context, in *VkTokenInfo, opts ...grpc.CallOption) (*Error, error)
	RemoveVkToken(ctx context.Context, in *VkTokenInfo, opts ...grpc.CallOption) (*Error, error)
}

type authClient struct {
	cc grpc.ClientConnInterface
}

func NewAuthClient(cc grpc.ClientConnInterface) AuthClient {
	return &authClient{cc}
}

func (c *authClient) CheckUserCredentials(ctx context.Context, in *UserAuth, opts ...grpc.CallOption) (*Error, error) {
	out := new(Error)
	err := c.cc.Invoke(ctx, "/auth.Auth/CheckUserCredentials", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authClient) AddCookieInfo(ctx context.Context, in *CookieInfo, opts ...grpc.CallOption) (*Error, error) {
	out := new(Error)
	err := c.cc.Invoke(ctx, "/auth.Auth/AddCookieInfo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authClient) SearchByValue(ctx context.Context, in *CookieValue, opts ...grpc.CallOption) (*CookieInfo, error) {
	out := new(CookieInfo)
	err := c.cc.Invoke(ctx, "/auth.Auth/SearchByValue", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authClient) SearchByUserID(ctx context.Context, in *UserID, opts ...grpc.CallOption) (*CookieInfo, error) {
	out := new(CookieInfo)
	err := c.cc.Invoke(ctx, "/auth.Auth/SearchByUserID", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authClient) RemoveCookie(ctx context.Context, in *CookieInfo, opts ...grpc.CallOption) (*Error, error) {
	out := new(Error)
	err := c.cc.Invoke(ctx, "/auth.Auth/RemoveCookie", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authClient) CheckUserByVkToken(ctx context.Context, in *VkToken, opts ...grpc.CallOption) (*UserID, error) {
	out := new(UserID)
	err := c.cc.Invoke(ctx, "/auth.Auth/CheckUserByVkToken", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authClient) AddVkToken(ctx context.Context, in *VkTokenInfo, opts ...grpc.CallOption) (*Error, error) {
	out := new(Error)
	err := c.cc.Invoke(ctx, "/auth.Auth/AddVkToken", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authClient) RemoveVkToken(ctx context.Context, in *VkTokenInfo, opts ...grpc.CallOption) (*Error, error) {
	out := new(Error)
	err := c.cc.Invoke(ctx, "/auth.Auth/RemoveVkToken", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AuthServer is the server API for Auth service.
type AuthServer interface {
	CheckUserCredentials(context.Context, *UserAuth) (*Error, error)
	AddCookieInfo(context.Context, *CookieInfo) (*Error, error)
	SearchByValue(context.Context, *CookieValue) (*CookieInfo, error)
	SearchByUserID(context.Context, *UserID) (*CookieInfo, error)
	RemoveCookie(context.Context, *CookieInfo) (*Error, error)
	CheckUserByVkToken(context.Context, *VkToken) (*UserID, error)
	AddVkToken(context.Context, *VkTokenInfo) (*Error, error)
	RemoveVkToken(context.Context, *VkTokenInfo) (*Error, error)
}

// UnimplementedAuthServer can be embedded to have forward compatible implementations.
type UnimplementedAuthServer struct {
}

func (*UnimplementedAuthServer) CheckUserCredentials(context.Context, *UserAuth) (*Error, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckUserCredentials not implemented")
}
func (*UnimplementedAuthServer) AddCookieInfo(context.Context, *CookieInfo) (*Error, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddCookieInfo not implemented")
}
func (*UnimplementedAuthServer) SearchByValue(context.Context, *CookieValue) (*CookieInfo, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SearchByValue not implemented")
}
func (*UnimplementedAuthServer) SearchByUserID(context.Context, *UserID) (*CookieInfo, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SearchByUserID not implemented")
}
func (*UnimplementedAuthServer) RemoveCookie(context.Context, *CookieInfo) (*Error, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveCookie not implemented")
}
func (*UnimplementedAuthServer) CheckUserByVkToken(context.Context, *VkToken) (*UserID, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckUserByVkToken not implemented")
}
func (*UnimplementedAuthServer) AddVkToken(context.Context, *VkTokenInfo) (*Error, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddVkToken not implemented")
}
func (*UnimplementedAuthServer) RemoveVkToken(context.Context, *VkTokenInfo) (*Error, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveVkToken not implemented")
}

func RegisterAuthServer(s *grpc.Server, srv AuthServer) {
	s.RegisterService(&_Auth_serviceDesc, srv)
}

func _Auth_CheckUserCredentials_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UserAuth)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServer).CheckUserCredentials(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/auth.Auth/CheckUserCredentials",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServer).CheckUserCredentials(ctx, req.(*UserAuth))
	}
	return interceptor(ctx, in, info, handler)
}

func _Auth_AddCookieInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CookieInfo)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServer).AddCookieInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/auth.Auth/AddCookieInfo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServer).AddCookieInfo(ctx, req.(*CookieInfo))
	}
	return interceptor(ctx, in, info, handler)
}

func _Auth_SearchByValue_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CookieValue)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServer).SearchByValue(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/auth.Auth/SearchByValue",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServer).SearchByValue(ctx, req.(*CookieValue))
	}
	return interceptor(ctx, in, info, handler)
}

func _Auth_SearchByUserID_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UserID)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServer).SearchByUserID(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/auth.Auth/SearchByUserID",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServer).SearchByUserID(ctx, req.(*UserID))
	}
	return interceptor(ctx, in, info, handler)
}

func _Auth_RemoveCookie_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CookieInfo)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServer).RemoveCookie(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/auth.Auth/RemoveCookie",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServer).RemoveCookie(ctx, req.(*CookieInfo))
	}
	return interceptor(ctx, in, info, handler)
}

func _Auth_CheckUserByVkToken_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(VkToken)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServer).CheckUserByVkToken(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/auth.Auth/CheckUserByVkToken",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServer).CheckUserByVkToken(ctx, req.(*VkToken))
	}
	return interceptor(ctx, in, info, handler)
}

func _Auth_AddVkToken_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(VkTokenInfo)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServer).AddVkToken(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/auth.Auth/AddVkToken",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServer).AddVkToken(ctx, req.(*VkTokenInfo))
	}
	return interceptor(ctx, in, info, handler)
}

func _Auth_RemoveVkToken_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(VkTokenInfo)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServer).RemoveVkToken(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/auth.Auth/RemoveVkToken",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServer).RemoveVkToken(ctx, req.(*VkTokenInfo))
	}
	return interceptor(ctx, in, info, handler)
}

var _Auth_serviceDesc = grpc.ServiceDesc{
	ServiceName: "auth.Auth",
	HandlerType: (*AuthServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CheckUserCredentials",
			Handler:    _Auth_CheckUserCredentials_Handler,
		},
		{
			MethodName: "AddCookieInfo",
			Handler:    _Auth_AddCookieInfo_Handler,
		},
		{
			MethodName: "SearchByValue",
			Handler:    _Auth_SearchByValue_Handler,
		},
		{
			MethodName: "SearchByUserID",
			Handler:    _Auth_SearchByUserID_Handler,
		},
		{
			MethodName: "RemoveCookie",
			Handler:    _Auth_RemoveCookie_Handler,
		},
		{
			MethodName: "CheckUserByVkToken",
			Handler:    _Auth_CheckUserByVkToken_Handler,
		},
		{
			MethodName: "AddVkToken",
			Handler:    _Auth_AddVkToken_Handler,
		},
		{
			MethodName: "RemoveVkToken",
			Handler:    _Auth_RemoveVkToken_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "auth.proto",
}
