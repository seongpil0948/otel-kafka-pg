// Copyright 2019, OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v5.29.3
// source: opentelemetry/proto/resource/v1/resource.proto

package v1

import (
	v1 "go.opentelemetry.io/proto/otlp/common/v1"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Resource information.
type Resource struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Set of attributes that describe the resource.
	// Attribute keys MUST be unique (it is not allowed to have more than one
	// attribute with the same key).
	Attributes []*v1.KeyValue `protobuf:"bytes,1,rep,name=attributes,proto3" json:"attributes,omitempty"`
	// dropped_attributes_count is the number of dropped attributes. If the value is 0, then
	// no attributes were dropped.
	DroppedAttributesCount uint32 `protobuf:"varint,2,opt,name=dropped_attributes_count,json=droppedAttributesCount,proto3" json:"dropped_attributes_count,omitempty"`
	// Set of entities that participate in this Resource.
	//
	// Note: keys in the references MUST exist in attributes of this message.
	//
	// Status: [Development]
	EntityRefs    []*v1.EntityRef `protobuf:"bytes,3,rep,name=entity_refs,json=entityRefs,proto3" json:"entity_refs,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Resource) Reset() {
	*x = Resource{}
	mi := &file_opentelemetry_proto_resource_v1_resource_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Resource) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Resource) ProtoMessage() {}

func (x *Resource) ProtoReflect() protoreflect.Message {
	mi := &file_opentelemetry_proto_resource_v1_resource_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Resource.ProtoReflect.Descriptor instead.
func (*Resource) Descriptor() ([]byte, []int) {
	return file_opentelemetry_proto_resource_v1_resource_proto_rawDescGZIP(), []int{0}
}

func (x *Resource) GetAttributes() []*v1.KeyValue {
	if x != nil {
		return x.Attributes
	}
	return nil
}

func (x *Resource) GetDroppedAttributesCount() uint32 {
	if x != nil {
		return x.DroppedAttributesCount
	}
	return 0
}

func (x *Resource) GetEntityRefs() []*v1.EntityRef {
	if x != nil {
		return x.EntityRefs
	}
	return nil
}

var File_opentelemetry_proto_resource_v1_resource_proto protoreflect.FileDescriptor

const file_opentelemetry_proto_resource_v1_resource_proto_rawDesc = "" +
	"\n" +
	".opentelemetry/proto/resource/v1/resource.proto\x12\x1fopentelemetry.proto.resource.v1\x1a*opentelemetry/proto/common/v1/common.proto\"\xd8\x01\n" +
	"\bResource\x12G\n" +
	"\n" +
	"attributes\x18\x01 \x03(\v2'.opentelemetry.proto.common.v1.KeyValueR\n" +
	"attributes\x128\n" +
	"\x18dropped_attributes_count\x18\x02 \x01(\rR\x16droppedAttributesCount\x12I\n" +
	"\ventity_refs\x18\x03 \x03(\v2(.opentelemetry.proto.common.v1.EntityRefR\n" +
	"entityRefsB\x83\x01\n" +
	"\"io.opentelemetry.proto.resource.v1B\rResourceProtoP\x01Z*go.opentelemetry.io/proto/otlp/resource/v1\xaa\x02\x1fOpenTelemetry.Proto.Resource.V1b\x06proto3"

var (
	file_opentelemetry_proto_resource_v1_resource_proto_rawDescOnce sync.Once
	file_opentelemetry_proto_resource_v1_resource_proto_rawDescData []byte
)

func file_opentelemetry_proto_resource_v1_resource_proto_rawDescGZIP() []byte {
	file_opentelemetry_proto_resource_v1_resource_proto_rawDescOnce.Do(func() {
		file_opentelemetry_proto_resource_v1_resource_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_opentelemetry_proto_resource_v1_resource_proto_rawDesc), len(file_opentelemetry_proto_resource_v1_resource_proto_rawDesc)))
	})
	return file_opentelemetry_proto_resource_v1_resource_proto_rawDescData
}

var file_opentelemetry_proto_resource_v1_resource_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_opentelemetry_proto_resource_v1_resource_proto_goTypes = []any{
	(*Resource)(nil),     // 0: opentelemetry.proto.resource.v1.Resource
	(*v1.KeyValue)(nil),  // 1: opentelemetry.proto.common.v1.KeyValue
	(*v1.EntityRef)(nil), // 2: opentelemetry.proto.common.v1.EntityRef
}
var file_opentelemetry_proto_resource_v1_resource_proto_depIdxs = []int32{
	1, // 0: opentelemetry.proto.resource.v1.Resource.attributes:type_name -> opentelemetry.proto.common.v1.KeyValue
	2, // 1: opentelemetry.proto.resource.v1.Resource.entity_refs:type_name -> opentelemetry.proto.common.v1.EntityRef
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_opentelemetry_proto_resource_v1_resource_proto_init() }
func file_opentelemetry_proto_resource_v1_resource_proto_init() {
	if File_opentelemetry_proto_resource_v1_resource_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_opentelemetry_proto_resource_v1_resource_proto_rawDesc), len(file_opentelemetry_proto_resource_v1_resource_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_opentelemetry_proto_resource_v1_resource_proto_goTypes,
		DependencyIndexes: file_opentelemetry_proto_resource_v1_resource_proto_depIdxs,
		MessageInfos:      file_opentelemetry_proto_resource_v1_resource_proto_msgTypes,
	}.Build()
	File_opentelemetry_proto_resource_v1_resource_proto = out.File
	file_opentelemetry_proto_resource_v1_resource_proto_goTypes = nil
	file_opentelemetry_proto_resource_v1_resource_proto_depIdxs = nil
}
