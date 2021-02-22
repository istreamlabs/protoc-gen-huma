package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"regexp"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/istreamlabs/protoc-gen-huma/annotation"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"

	"github.com/danielgtaylor/casing"
	"github.com/flosch/pongo2"
)

// spaceRegex is used to collapse/simplify consecutive whitespace characters.
var spaceRegex = regexp.MustCompile(`\s+`)

// registry of Protobuf types. These can be one of:
// - descriptorpb.EnumDescriptorProto
// - descriptorpb.DiscriptorProto
// This is used to load all types and provide a quick lookup for type info
// across proto packages.
var registry map[string]interface{} = map[string]interface{}{}

// goCase returns camel cased names with capitalized initialisms that pass
// the Go linter.
func goCase(values ...string) string {
	return casing.Camel(strings.Join(values, "_"), strings.ToLower, casing.Initialism)
}

// getComments for a message, field, enum, etc. Comments in protobuf use a path
// of field numbers, e.g. [4, 2, 1] would mean FileDescriptor field 4 which
// points to a list of message descriptors, the 2nd message in that list and
// the first field in the message Descriptor type. Each path can have leading
// and trailing comments. Yes, this is a batshit crazy way to do things.
func getComments(tFile *File, path []int32) string {
	for _, loc := range tFile.Proto.SourceCodeInfo.Location {
		if reflect.DeepEqual(loc.Path, path) && loc.LeadingComments != nil {
			// Replace double quotes, used for struct tags.
			comment := strings.Replace(*loc.LeadingComments, `"`, "'", -1)

			// Replace backticks, used for struct tags.
			comment = strings.Replace(comment, "`", "'", -1)

			// Replace newlines/tabs/multi-spaces for custom flow.
			comment = spaceRegex.ReplaceAllString(comment, " ")

			// Trim leading/trailing spaces.
			comment = strings.TrimSpace(comment)

			return comment
		}
	}
	return ""
}

// addEnums adds all the proto enum information to the file. Enums can come from
// the top-level file or be embedded inside nested messages.
func addEnums(tFile *File, prefix string, path []int32, enums []*descriptor.EnumDescriptorProto) {
	// If nested, prefix will be set with the outer name. We append to it below.
	p := casing.Join(strings.Split(prefix, " "), "_", casing.Identity)
	if p != "" {
		p += "_"
	}

	for _, e := range enums {
		tEnum := Enum{
			Name:        goCase(prefix + " " + e.GetName()),
			ProtoGoName: p + casing.Camel(e.GetName(), casing.Identity),
			Values:      []EnumValue{},
		}

		for _, v := range e.GetValue() {
			if e := proto.GetExtension(v.GetOptions(), annotation.E_Exclude).(bool); e {
				// Skip this enum value!
				continue
			}

			tEnum.Values = append(tEnum.Values, EnumValue{
				Name:  goCase(v.GetName()),
				Label: v.GetName(),
				Value: v.GetNumber(),
			})
		}

		if !tFile.KnownMap[tEnum.Name] {
			tFile.Enums = append(tFile.Enums, tEnum)
		}
	}
}

// getType returns the Go type, protobuf-generated Go type, whether the type is
// a primitive or not, and which enum corresponds to the type if any.
func getType(tFile *File, prefix string, f *descriptorpb.FieldDescriptorProto) (string, string, bool, *Enum) {
	t := ""
	pt := ""
	primitive := true
	var enum *Enum

	// First, convert all the known protobuf types to their Go equivalents.
	switch *f.Type {
	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		t = "bool"
	case descriptor.FieldDescriptorProto_TYPE_INT32, descriptor.FieldDescriptorProto_TYPE_SINT32, descriptor.FieldDescriptorProto_TYPE_SFIXED32:
		t = "int32"
	case descriptor.FieldDescriptorProto_TYPE_INT64, descriptor.FieldDescriptorProto_TYPE_SINT64, descriptor.FieldDescriptorProto_TYPE_SFIXED64:
		t = "int64"
	case descriptor.FieldDescriptorProto_TYPE_UINT32, descriptor.FieldDescriptorProto_TYPE_FIXED32:
		t = "uint32"
	case descriptor.FieldDescriptorProto_TYPE_UINT64, descriptor.FieldDescriptorProto_TYPE_FIXED64:
		t = "uint64"
	case descriptor.FieldDescriptorProto_TYPE_FLOAT:
		t = "float32"
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		t = "float64"
	case descriptor.FieldDescriptorProto_TYPE_STRING:
		t = "string"
	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		t = "[]byte"
	case descriptor.FieldDescriptorProto_TYPE_ENUM:
		parts := strings.Split(*f.TypeName, ".")
		prefix := ""
		if parts[1] != tFile.PackageName {
			prefix += parts[1] + "."
		}
		t = prefix + goCase(strings.Join(parts[2:], "_"))
		pt = parts[1] + "." + strings.Join(parts[2:], "_")
		primitive = false

		for _, e := range tFile.Enums {
			if e.Name == t {
				enum = &e
				break
			}
		}
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		if *f.TypeName == ".google.protobuf.Timestamp" {
			tFile.Imports["time"] = true
			tFile.Imports["google.golang.org/protobuf/types/known/timestamppb"] = true

			return "*time.Time", "", false, nil
		}

		// Special case: map types generate an intermediary message type that
		// represents an entry in the map as a repeated message. We only care
		// about the value type and assume string keys here.
		if d, ok := registry[*f.TypeName]; ok {
			if proto, ok := d.(*descriptorpb.DescriptorProto); ok {
				if proto.Options != nil && proto.Options.MapEntry != nil && *proto.Options.MapEntry {
					// Field 0 = key, field 1 = value for every generated message.
					t, pt, primitive, enum = getType(tFile, prefix, proto.Field[1])
					t = "map[string]" + t
					pt = "map[string]" + pt
					return t, pt, primitive, enum
				}
			}
		}

		parts := strings.Split(*f.TypeName, ".")

		prefix := "*"
		if parts[1] != tFile.PackageName {
			prefix += parts[1] + "."
		}
		t = prefix + goCase(parts[2:]...)
		pt = "*" + parts[1] + "." + strings.Join(parts[2:], "_")
		primitive = false
	default:
		spew.Fdump(os.Stderr, f)
		panic("Unknown type")
	}

	if pt == "" {
		pt = t
	}

	return t, pt, primitive, enum
}

// newField makes a field description from a protobuf field.
func newField(tFile *File, protoMessage *descriptorpb.DescriptorProto, fieldPath []int32, protoField *descriptorpb.FieldDescriptorProto) *Field {
	name := goCase(protoField.GetName())
	jsName := casing.Snake(protoField.GetJsonName())

	if s := proto.GetExtension(protoField.GetOptions(), annotation.E_Name).(string); s != "" {
		name = s
		jsName = casing.Snake(s)
	}
	if s := proto.GetExtension(protoField.GetOptions(), annotation.E_Json).(string); s != "" {
		jsName = s
	}

	example := ""
	if e := proto.GetExtension(protoField.GetOptions(), annotation.E_Example).(string); e != "" {
		example = e
	}

	f := &Field{
		Name:        name,
		ProtoGoName: casing.Camel(protoField.GetName()),
		JSONName:    jsName,
		Comment:     getComments(tFile, fieldPath),
		Example:     example,
	}

	f.GoType, f.ProtoGoType, f.IsPrimitive, f.Enum = getType(tFile, "", protoField)
	f.IsMap = strings.HasPrefix(f.GoType, "map[")

	if !f.IsMap && protoField.Label != nil && *protoField.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED {
		// This is a slice of values, so update the types.
		f.IsRepeated = true
		f.GoType = "[]" + f.GoType
		f.ProtoGoType = "[]" + f.ProtoGoType
	}

	if protoField.OneofIndex != nil {
		// This field is part of a "one-of" group. In generated Go code this turns
		// into a single field, so we set the one-of name for all the fields in the
		// group to that one field name. The `ProtoGoType` stays the same as each
		// item in the group can still be a unique type, just *where* we set it
		// in the generated Go struct changes.
		f.OneOf = casing.Camel(protoMessage.OneofDecl[int(*protoField.OneofIndex)].GetName())
	}

	convertValidation(protoField, f)

	return f
}

// traverse a set of messages recursively.
func traverse(tFile *File, prefix string, path []int32, items []*descriptorpb.DescriptorProto) {
	for i, msg := range items {
		if msg.Options != nil && msg.Options.MapEntry != nil && *msg.Options.MapEntry {
			// Skip generated map entry objects.
			continue
		}

		msgPath := append(append([]int32{}, path...), int32(i))
		p := casing.Camel(prefix, casing.Identity)
		if p != "" {
			p += "_"
		}
		tMsg := Message{
			Name:        goCase(prefix + " " + msg.GetName()),
			ProtoGoName: p + casing.Camel(msg.GetName(), casing.Identity),
			Fields:      []*Field{},
			OneOfs:      map[string][]*Field{},
			Comment:     getComments(tFile, msgPath),
		}

		// Parse nested types first since the fields may reference them.
		if len(msg.EnumType) > 0 {
			newPath := append(append([]int32{}, msgPath...), 4)
			addEnums(tFile, prefix+" "+msg.GetName(), newPath, msg.EnumType)
		}

		if len(msg.NestedType) > 0 {
			newPath := append(append([]int32{}, msgPath...), 3)
			traverse(tFile, prefix+" "+msg.GetName(), newPath, msg.NestedType)
		}

		for j, f := range msg.Field {
			// Only expose public fields!
			if proto.GetExtension(f.GetOptions(), annotation.E_Public).(bool) || os.Getenv("ALL_PUBLIC") != "" {
				fieldPath := append(append([]int32{}, msgPath...), 2, int32(j))
				tField := newField(tFile, msg, fieldPath, f)

				if tField.OneOf != "" {
					// One-of fields have some extra rules and require some additional
					// packages.
					tFile.Imports["net/http"] = true
					tFile.Imports["reflect"] = true
					tFile.Imports["strings"] = true
					tFile.Imports["github.com/istreamlabs/huma"] = true
					tMsg.OneOfs[tField.OneOf] = append(tMsg.OneOfs[tField.OneOf], tField)
				}

				// Add the new field to the message type.
				tMsg.Fields = append(tMsg.Fields, tField)
			}
		}

		// All fields are loaded, document one-ofs so users know which fields
		// are mutually exclusive since we handle this with custom Huma validation
		// logic instead of JSON Schema.
		for _, fields := range tMsg.OneOfs {
			names := []string{}
			for _, f := range fields {
				names = append(names, f.JSONName)
			}
			for _, f := range fields {
				if f.Comment != "" {
					f.Comment += " "
				}
				f.Comment += "Only one of ['" + strings.Join(names, "', '") + "'] may be set."
			}
		}

		if !tFile.KnownMap[tMsg.Name] {
			tFile.KnownMap[tMsg.Name] = true
			tFile.Messages = append(tFile.Messages, tMsg)
		}
	}
}

// buildRegistry makes a map of fully qualified type names to their proto
// descriptors.
func buildRegistry(prefix string, messages []*descriptorpb.DescriptorProto, enums []*descriptorpb.EnumDescriptorProto) {
	for _, msg := range messages {
		registry[prefix+"."+*msg.Name] = msg
		buildRegistry(prefix+"."+*msg.Name, msg.NestedType, msg.EnumType)
	}
}

func run(input []byte) []byte {
	// Protoc passes pluginpb.CodeGeneratorRequest in via stdin
	// marshalled with Protobuf.
	var req pluginpb.CodeGeneratorRequest
	proto.Unmarshal(input, &req)

	// Initialise our plugin with default options.
	opts := protogen.Options{}
	plugin, err := opts.New(&req)
	if err != nil {
		panic(err)
	}

	// Create a map of files we've been explicitly asked to generate for fast
	// lookups. The `plugin.Files` used below also contains any imports and we
	// don't want to generate those in the output.
	filesToGen := map[string]bool{}
	for _, file := range plugin.Request.FileToGenerate {
		filesToGen[file] = true
	}

	// Generate a type registry map of fully-qualified type names (e.g.
	// `.google.protobuf.duration`) to the corresponding message or enum
	// descriptor for lookups later.
	for _, file := range plugin.Files {
		buildRegistry("."+*file.Proto.Package, file.Proto.MessageType, file.Proto.EnumType)
	}

	// Protoc passes a slice of File structs for us to process
	for _, file := range plugin.Files {
		if !filesToGen[file.Desc.Path()] {
			// Skip anything that wasn't explicitly asked for (i.e. built-ins and dependencies).
			continue
		}

		tFile := File{
			Proto:         file.Proto,
			PackageName:   fmt.Sprintf("%s", file.GoPackageName),
			Imports:       map[string]bool{string(file.GoImportPath): true},
			ProtoGoImport: *file.Proto.Options.GoPackage,
			KnownMap:      map[string]bool{},
			Messages:      []Message{},
		}

		// Add all the public types from the file. The magic numbers below are from
		// the FileDescriptorProto message, see:
		// https://github.com/protocolbuffers/protobuf/blob/master/src/google/protobuf/descriptor.proto#L75-L76
		addEnums(&tFile, "", []int32{5}, file.Proto.EnumType)
		traverse(&tFile, "", []int32{4}, file.Proto.MessageType)

		// Only output the file if it has actual public stuff in it.
		if len(tFile.Messages) > 0 || len(tFile.Enums) > 0 {
			p := file.Desc.Path()
			filename := p[:len(p)-len(path.Ext(p))] + ".huma.go"
			file := plugin.NewGeneratedFile(filename, ".")
			if err := humaTemplate.ExecuteWriter(pongo2.Context{"file": tFile}, file); err != nil {
				panic(err)
			}
		}
	}

	// Generate a response from our plugin and marshall as protobuf
	stdout := plugin.Response()
	out, err := proto.Marshal(stdout)
	if err != nil {
		panic(err)
	}
	return out
}

func main() {
	// Disable HTML escaping, we are generating Go code!
	pongo2.SetAutoescape(false)

	// Protoc passes our input data via stdin.
	input, _ := ioutil.ReadAll(os.Stdin)

	if os.Getenv("DUMP_REQUEST") != "" {
		ioutil.WriteFile("request.pb", input, os.ModePerm)
	}

	out := run(input)

	// Write the response to stdout, to be picked up by protoc.
	os.Stdout.Write(out)
}
