package main

import "google.golang.org/protobuf/types/descriptorpb"

// EnumValue represents one protobuf enum value.
type EnumValue struct {
	// Name is the Huma name for the enum value.
	Name string

	// Label is the protobuf label for the enum value.
	Label string

	// Value is the protobuf integer assigned to the enum value.
	Value int32
}

// Enum represents a protobuf enum.
type Enum struct {
	// Name is the Huma name for the enum type.
	Name string

	// ProtoGoName is the protobuf-generated Go type name for the enum.
	ProtoGoName string

	// Values represents all the defined values for the enum.
	Values []EnumValue

	// Comment is the leading comment for the enum type, if any.
	Comment string
}

// Field represents a protobuf field within a message.
type Field struct {
	// Name is the Huma name for the field.
	Name string

	// ProtoGoName is the protobuf-generated Go name for the field.
	ProtoGoName string

	// JSONName is the snake-cased JSON name for the field.
	JSONName string

	// GoType is the type of the field, e.g. string or int64
	GoType string

	// ProtoGoType is the protobuf-generated Go type for the field. It may equal
	// the GoType for primitives but will differ for enums & messages.
	ProtoGoType string

	// IsMap is true if the field is a map type.
	IsMap bool

	// IsPrimitive is true if the field is a primitive, e.g. bool, int32, float32,
	// string, etc.
	IsPrimitive bool

	// IsRepeated is true if the field is an array type.
	IsRepeated bool

	// OneOf is set to the one-of group name if the field is part of a one-of
	// group, otherwise it is blank.
	OneOf string

	// Comment is the leading comment for the field, if any.
	Comment string

	// Enum is non-nil if this field is an enum type.
	Enum *Enum

	// Validation contains validation rules for this field.
	Validation Validation
}

// Message represents a protobuf message type whithin a file.
type Message struct {
	// Name is the Huma name for this message type.
	Name string

	// ProtoGoName is the protobuf-generated Go name for this type.
	ProtoGoName string

	// Fields is a slice of field definitions in the message.
	Fields []*Field

	// OneOfs is a map of one-of names to fields.
	OneOfs map[string][]*Field

	// Comment is the leading comment for the message, if any.
	Comment string
}

// File represents a protobuf file.
type File struct {
	// Proto is the protobuf file descriptor for this file.
	Proto *descriptorpb.FileDescriptorProto

	// PackageName is the Go package name.
	PackageName string

	// ProtoGoImport is the import path to the protobuf-generated Go output.
	ProtoGoImport string

	// Imports is a list of Go imports for the file, based on which types and
	// features are used.
	Imports map[string]bool

	// KnownMap is used to keep track of which enums and messages have been seen
	// before so we don't get duplicate definitions.
	KnownMap map[string]bool

	// Messages is a slice of message definitions in the file.
	Messages []Message

	// Enums is a slice of enum definitions in the file.
	Enums []Enum
}
