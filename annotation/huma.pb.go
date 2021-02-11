// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.14.0
// source: huma.proto

package annotation

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	descriptorpb "google.golang.org/protobuf/types/descriptorpb"
	reflect "reflect"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

var file_huma_proto_extTypes = []protoimpl.ExtensionInfo{
	{
		ExtendedType:  (*descriptorpb.FieldOptions)(nil),
		ExtensionType: (*bool)(nil),
		Field:         84841,
		Name:          "huma.public",
		Tag:           "varint,84841,opt,name=public",
		Filename:      "huma.proto",
	},
	{
		ExtendedType:  (*descriptorpb.FieldOptions)(nil),
		ExtensionType: (*bool)(nil),
		Field:         84842,
		Name:          "huma.readOnly",
		Tag:           "varint,84842,opt,name=readOnly",
		Filename:      "huma.proto",
	},
	{
		ExtendedType:  (*descriptorpb.FieldOptions)(nil),
		ExtensionType: (*string)(nil),
		Field:         84843,
		Name:          "huma.name",
		Tag:           "bytes,84843,opt,name=name",
		Filename:      "huma.proto",
	},
	{
		ExtendedType:  (*descriptorpb.FieldOptions)(nil),
		ExtensionType: (*string)(nil),
		Field:         84844,
		Name:          "huma.json",
		Tag:           "bytes,84844,opt,name=json",
		Filename:      "huma.proto",
	},
}

// Extension fields to descriptorpb.FieldOptions.
var (
	// Public marks that a field should be included in the generated Huma model.
	//
	// optional bool public = 84841;
	E_Public = &file_huma_proto_extTypes[0]
	// ReadOnly marks that a field is set by the server. The client can only
	// read its value, e.g. resource creation date.
	//
	// optional bool readOnly = 84842;
	E_ReadOnly = &file_huma_proto_extTypes[1]
	// Name specifies the Huma Go field name. For example, a field might be cased
	// as `Mp2T` but should be `MP2T` because it is a non-common initialism.
	// Setting this field also updates the JSON name unless it has also been
	// overridden.
	//
	// optional string name = 84843;
	E_Name = &file_huma_proto_extTypes[2]
	// JSON specifies the Huma field's JSON name. Usually this is derived from
	// the field's name, but this option allows you to override it. For example,
	// a field named `MP2T` might become `mp2_t` but should be `mp2t`.
	//
	// optional string json = 84844;
	E_Json = &file_huma_proto_extTypes[3]
)

var File_huma_proto protoreflect.FileDescriptor

var file_huma_proto_rawDesc = []byte{
	0x0a, 0x0a, 0x68, 0x75, 0x6d, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x04, 0x68, 0x75,
	0x6d, 0x61, 0x1a, 0x20, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2f, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x6f, 0x72, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x3a, 0x3a, 0x0a, 0x06, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x12, 0x1d,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xe9, 0x96,
	0x05, 0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x88, 0x01, 0x01,
	0x3a, 0x3e, 0x0a, 0x08, 0x72, 0x65, 0x61, 0x64, 0x4f, 0x6e, 0x6c, 0x79, 0x12, 0x1d, 0x2e, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x46,
	0x69, 0x65, 0x6c, 0x64, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xea, 0x96, 0x05, 0x20,
	0x01, 0x28, 0x08, 0x52, 0x08, 0x72, 0x65, 0x61, 0x64, 0x4f, 0x6e, 0x6c, 0x79, 0x88, 0x01, 0x01,
	0x3a, 0x36, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x1d, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x46, 0x69, 0x65, 0x6c, 0x64,
	0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xeb, 0x96, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x6e, 0x61, 0x6d, 0x65, 0x88, 0x01, 0x01, 0x3a, 0x36, 0x0a, 0x04, 0x6a, 0x73, 0x6f, 0x6e,
	0x12, 0x1d, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2e, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18,
	0xec, 0x96, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6a, 0x73, 0x6f, 0x6e, 0x88, 0x01, 0x01,
	0x42, 0x33, 0x5a, 0x31, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x69,
	0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x6c, 0x61, 0x62, 0x73, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x63, 0x2d, 0x67, 0x65, 0x6e, 0x2d, 0x68, 0x75, 0x6d, 0x61, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var file_huma_proto_goTypes = []interface{}{
	(*descriptorpb.FieldOptions)(nil), // 0: google.protobuf.FieldOptions
}
var file_huma_proto_depIdxs = []int32{
	0, // 0: huma.public:extendee -> google.protobuf.FieldOptions
	0, // 1: huma.readOnly:extendee -> google.protobuf.FieldOptions
	0, // 2: huma.name:extendee -> google.protobuf.FieldOptions
	0, // 3: huma.json:extendee -> google.protobuf.FieldOptions
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	0, // [0:4] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_huma_proto_init() }
func file_huma_proto_init() {
	if File_huma_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_huma_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 4,
			NumServices:   0,
		},
		GoTypes:           file_huma_proto_goTypes,
		DependencyIndexes: file_huma_proto_depIdxs,
		ExtensionInfos:    file_huma_proto_extTypes,
	}.Build()
	File_huma_proto = out.File
	file_huma_proto_rawDesc = nil
	file_huma_proto_goTypes = nil
	file_huma_proto_depIdxs = nil
}
