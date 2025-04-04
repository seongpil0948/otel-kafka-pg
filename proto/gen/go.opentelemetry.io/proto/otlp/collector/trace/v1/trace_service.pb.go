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
// source: opentelemetry/proto/collector/trace/v1/trace_service.proto

package v1

import (
	v1 "go.opentelemetry.io/proto/otlp/trace/v1"
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

type ExportTraceServiceRequest struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// An array of ResourceSpans.
	// For data coming from a single resource this array will typically contain one
	// element. Intermediary nodes (such as OpenTelemetry Collector) that receive
	// data from multiple origins typically batch the data before forwarding further and
	// in that case this array will contain multiple elements.
	ResourceSpans []*v1.ResourceSpans `protobuf:"bytes,1,rep,name=resource_spans,json=resourceSpans,proto3" json:"resource_spans,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ExportTraceServiceRequest) Reset() {
	*x = ExportTraceServiceRequest{}
	mi := &file_opentelemetry_proto_collector_trace_v1_trace_service_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ExportTraceServiceRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ExportTraceServiceRequest) ProtoMessage() {}

func (x *ExportTraceServiceRequest) ProtoReflect() protoreflect.Message {
	mi := &file_opentelemetry_proto_collector_trace_v1_trace_service_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ExportTraceServiceRequest.ProtoReflect.Descriptor instead.
func (*ExportTraceServiceRequest) Descriptor() ([]byte, []int) {
	return file_opentelemetry_proto_collector_trace_v1_trace_service_proto_rawDescGZIP(), []int{0}
}

func (x *ExportTraceServiceRequest) GetResourceSpans() []*v1.ResourceSpans {
	if x != nil {
		return x.ResourceSpans
	}
	return nil
}

type ExportTraceServiceResponse struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// The details of a partially successful export request.
	//
	// If the request is only partially accepted
	// (i.e. when the server accepts only parts of the data and rejects the rest)
	// the server MUST initialize the `partial_success` field and MUST
	// set the `rejected_<signal>` with the number of items it rejected.
	//
	// Servers MAY also make use of the `partial_success` field to convey
	// warnings/suggestions to senders even when the request was fully accepted.
	// In such cases, the `rejected_<signal>` MUST have a value of `0` and
	// the `error_message` MUST be non-empty.
	//
	// A `partial_success` message with an empty value (rejected_<signal> = 0 and
	// `error_message` = "") is equivalent to it not being set/present. Senders
	// SHOULD interpret it the same way as in the full success case.
	PartialSuccess *ExportTracePartialSuccess `protobuf:"bytes,1,opt,name=partial_success,json=partialSuccess,proto3" json:"partial_success,omitempty"`
	unknownFields  protoimpl.UnknownFields
	sizeCache      protoimpl.SizeCache
}

func (x *ExportTraceServiceResponse) Reset() {
	*x = ExportTraceServiceResponse{}
	mi := &file_opentelemetry_proto_collector_trace_v1_trace_service_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ExportTraceServiceResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ExportTraceServiceResponse) ProtoMessage() {}

func (x *ExportTraceServiceResponse) ProtoReflect() protoreflect.Message {
	mi := &file_opentelemetry_proto_collector_trace_v1_trace_service_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ExportTraceServiceResponse.ProtoReflect.Descriptor instead.
func (*ExportTraceServiceResponse) Descriptor() ([]byte, []int) {
	return file_opentelemetry_proto_collector_trace_v1_trace_service_proto_rawDescGZIP(), []int{1}
}

func (x *ExportTraceServiceResponse) GetPartialSuccess() *ExportTracePartialSuccess {
	if x != nil {
		return x.PartialSuccess
	}
	return nil
}

type ExportTracePartialSuccess struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// The number of rejected spans.
	//
	// A `rejected_<signal>` field holding a `0` value indicates that the
	// request was fully accepted.
	RejectedSpans int64 `protobuf:"varint,1,opt,name=rejected_spans,json=rejectedSpans,proto3" json:"rejected_spans,omitempty"`
	// A developer-facing human-readable message in English. It should be used
	// either to explain why the server rejected parts of the data during a partial
	// success or to convey warnings/suggestions during a full success. The message
	// should offer guidance on how users can address such issues.
	//
	// error_message is an optional field. An error_message with an empty value
	// is equivalent to it not being set.
	ErrorMessage  string `protobuf:"bytes,2,opt,name=error_message,json=errorMessage,proto3" json:"error_message,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ExportTracePartialSuccess) Reset() {
	*x = ExportTracePartialSuccess{}
	mi := &file_opentelemetry_proto_collector_trace_v1_trace_service_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ExportTracePartialSuccess) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ExportTracePartialSuccess) ProtoMessage() {}

func (x *ExportTracePartialSuccess) ProtoReflect() protoreflect.Message {
	mi := &file_opentelemetry_proto_collector_trace_v1_trace_service_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ExportTracePartialSuccess.ProtoReflect.Descriptor instead.
func (*ExportTracePartialSuccess) Descriptor() ([]byte, []int) {
	return file_opentelemetry_proto_collector_trace_v1_trace_service_proto_rawDescGZIP(), []int{2}
}

func (x *ExportTracePartialSuccess) GetRejectedSpans() int64 {
	if x != nil {
		return x.RejectedSpans
	}
	return 0
}

func (x *ExportTracePartialSuccess) GetErrorMessage() string {
	if x != nil {
		return x.ErrorMessage
	}
	return ""
}

var File_opentelemetry_proto_collector_trace_v1_trace_service_proto protoreflect.FileDescriptor

const file_opentelemetry_proto_collector_trace_v1_trace_service_proto_rawDesc = "" +
	"\n" +
	":opentelemetry/proto/collector/trace/v1/trace_service.proto\x12&opentelemetry.proto.collector.trace.v1\x1a(opentelemetry/proto/trace/v1/trace.proto\"o\n" +
	"\x19ExportTraceServiceRequest\x12R\n" +
	"\x0eresource_spans\x18\x01 \x03(\v2+.opentelemetry.proto.trace.v1.ResourceSpansR\rresourceSpans\"\x88\x01\n" +
	"\x1aExportTraceServiceResponse\x12j\n" +
	"\x0fpartial_success\x18\x01 \x01(\v2A.opentelemetry.proto.collector.trace.v1.ExportTracePartialSuccessR\x0epartialSuccess\"g\n" +
	"\x19ExportTracePartialSuccess\x12%\n" +
	"\x0erejected_spans\x18\x01 \x01(\x03R\rrejectedSpans\x12#\n" +
	"\rerror_message\x18\x02 \x01(\tR\ferrorMessage2\xa2\x01\n" +
	"\fTraceService\x12\x91\x01\n" +
	"\x06Export\x12A.opentelemetry.proto.collector.trace.v1.ExportTraceServiceRequest\x1aB.opentelemetry.proto.collector.trace.v1.ExportTraceServiceResponse\"\x00B\x9c\x01\n" +
	")io.opentelemetry.proto.collector.trace.v1B\x11TraceServiceProtoP\x01Z1go.opentelemetry.io/proto/otlp/collector/trace/v1\xaa\x02&OpenTelemetry.Proto.Collector.Trace.V1b\x06proto3"

var (
	file_opentelemetry_proto_collector_trace_v1_trace_service_proto_rawDescOnce sync.Once
	file_opentelemetry_proto_collector_trace_v1_trace_service_proto_rawDescData []byte
)

func file_opentelemetry_proto_collector_trace_v1_trace_service_proto_rawDescGZIP() []byte {
	file_opentelemetry_proto_collector_trace_v1_trace_service_proto_rawDescOnce.Do(func() {
		file_opentelemetry_proto_collector_trace_v1_trace_service_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_opentelemetry_proto_collector_trace_v1_trace_service_proto_rawDesc), len(file_opentelemetry_proto_collector_trace_v1_trace_service_proto_rawDesc)))
	})
	return file_opentelemetry_proto_collector_trace_v1_trace_service_proto_rawDescData
}

var file_opentelemetry_proto_collector_trace_v1_trace_service_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_opentelemetry_proto_collector_trace_v1_trace_service_proto_goTypes = []any{
	(*ExportTraceServiceRequest)(nil),  // 0: opentelemetry.proto.collector.trace.v1.ExportTraceServiceRequest
	(*ExportTraceServiceResponse)(nil), // 1: opentelemetry.proto.collector.trace.v1.ExportTraceServiceResponse
	(*ExportTracePartialSuccess)(nil),  // 2: opentelemetry.proto.collector.trace.v1.ExportTracePartialSuccess
	(*v1.ResourceSpans)(nil),           // 3: opentelemetry.proto.trace.v1.ResourceSpans
}
var file_opentelemetry_proto_collector_trace_v1_trace_service_proto_depIdxs = []int32{
	3, // 0: opentelemetry.proto.collector.trace.v1.ExportTraceServiceRequest.resource_spans:type_name -> opentelemetry.proto.trace.v1.ResourceSpans
	2, // 1: opentelemetry.proto.collector.trace.v1.ExportTraceServiceResponse.partial_success:type_name -> opentelemetry.proto.collector.trace.v1.ExportTracePartialSuccess
	0, // 2: opentelemetry.proto.collector.trace.v1.TraceService.Export:input_type -> opentelemetry.proto.collector.trace.v1.ExportTraceServiceRequest
	1, // 3: opentelemetry.proto.collector.trace.v1.TraceService.Export:output_type -> opentelemetry.proto.collector.trace.v1.ExportTraceServiceResponse
	3, // [3:4] is the sub-list for method output_type
	2, // [2:3] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_opentelemetry_proto_collector_trace_v1_trace_service_proto_init() }
func file_opentelemetry_proto_collector_trace_v1_trace_service_proto_init() {
	if File_opentelemetry_proto_collector_trace_v1_trace_service_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_opentelemetry_proto_collector_trace_v1_trace_service_proto_rawDesc), len(file_opentelemetry_proto_collector_trace_v1_trace_service_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_opentelemetry_proto_collector_trace_v1_trace_service_proto_goTypes,
		DependencyIndexes: file_opentelemetry_proto_collector_trace_v1_trace_service_proto_depIdxs,
		MessageInfos:      file_opentelemetry_proto_collector_trace_v1_trace_service_proto_msgTypes,
	}.Build()
	File_opentelemetry_proto_collector_trace_v1_trace_service_proto = out.File
	file_opentelemetry_proto_collector_trace_v1_trace_service_proto_goTypes = nil
	file_opentelemetry_proto_collector_trace_v1_trace_service_proto_depIdxs = nil
}
