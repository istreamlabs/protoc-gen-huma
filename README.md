# Huma Protocol Buffer Compiler Plugin

Generates Go code from protobuf files for use with [Huma](https://huma.rocks/). Makes it easy to expose an external HTTP API using Huma when one or more teams internally use protobuf (e.g. gRPC / Twirp).

Why use this?

- Prevent mistakes from hand-written conversion code
- Team and subject matter expert autonomy
- Simplified data model since it matches the internal one

## Assumptions

- You will write handlers by hand (this just generates data structures)
- Everything is private unless explicitly marked as public
- Map keys **must** be strings
- Everything is optional unless explicitly marked as required
- If you add validation, you use [protoc-gen-validate](https://github.com/envoyproxy/protoc-gen-validate)

## Features

The following protobuf features are supported:

- Comments
- Primitive fields
  - `bool`
  - `int32`, `int64`, `uint32`, `uint64`, `sint32`, `sint64`
  - `float`, `double`
  - `fixed32`, `fixed64`, `sfixed32`, `sfixed64`
  - `string`, `bytes`
- Enums & nested enums
- Messages & nested messages
- Well known types (only those listed below)
  - [`google.protobuf.Timestamp`](https://developers.google.com/protocol-buffers/docs/reference/google.protobuf#timestamp)
- Arrays of primitives, enums, and messages
- Maps, represented via Go `map[string]...`
- One-of fields
- Deprecated annotations
- Validation via protoc-gen-validate annotations
  - Message `required`
  - Common numerics `lt`, `lte`, `gt`, `gte`
  - String `min_len`, `max_len`, `pattern`, various formats like `uri-ref`
  - Arrays `min_items`, `max_items`, `unique`
  - Enum `not_in`

## Annotations

The following new annotations are supported when using the `annotation/huma.proto` import:

| Name       | Type   | Example                    | Description                                                                           |
| ---------- | ------ | -------------------------- | ------------------------------------------------------------------------------------- |
| `public`   | `bool` | `[(huma.public) = true]`   | Make a field public.                                                                  |
| `readOnly` | `bool` | `[(huma.readOnly) = true]` | Prevent writing to the field, useful for server-generated values, e.g. creation time. |

## Example

Here is an example showing what the input and output might look like:

```proto
syntax = "proto3";

package example;

import "annotation/huma.proto";

option go_package = ".;example";

// Message description...
message Message {
  // Some data goes here.
  string data = 1 [(huma.public) = true];
}
```

That will generate:

```go
package examplehuma

// Message description...
type Message struct {
  // Some data goes here.
  data string `json:"data,omitempty" doc:"Some data goes here."`
}

// FromProto converts a proto message to the Huma representation.
func (m *Message) FromProto(proto *example.Message) {
  m.data = proto.data
}

// ToProto converts a Huma representation to a proto message.
func (m *Message) ToProto(proto *example.Message) *example.Message {
  if proto == nil {
    proto = &example.Message{}
  }

  proto.data = m.data

  return proto
}
```

## Design

The Huma protobuf compiler plugin takes in a description of parsed proto files, processes each file to get a list of message and enum types, converts them to Huma representations, and renders out a corresponding file for each input proto using a template.

```
file1.proto \                            /-> process file1 -> file1.huma.go
file2.proto -> protoc -> protoc-gen-huma --> process file2 -> file2.huma.go
file3.proto /                            \-> process file3 -> file3.huma.go
```

High level pseudo-code:

- For each incoming proto file
  - Collect enums
    - Figure out Huma naming, treat as strings
  - Collect messages
    - For each message
      - Figure out Huma naming & type
      - For each _public_ field
        - Figure out Huma naming & type
        - Convert validation to Huma's JSON-Schema tags
      - Generate `FromProto` and `ToProto` converter methods
  - Write out `$BASENAME.huma.go`

## Implementation Details

### Comments

Comments in parsed protobuf files are stored in the [`FileDescriptorProto.SourceCodeInfo`](https://github.com/protocolbuffers/protobuf/blob/master/src/google/protobuf/descriptor.proto#L86) field. Each location in that list is defined as a path of _proto field numbers_ from the top-level `FileDescriptorProto` to the current enum/message/field. For example, a path like `[4, 1, 2, 5]` would mean the `FileDescriptorProto` field `4` (aka message types), message `1` in that array, field `2` of the message (e.g. maybe a nested message), then field `5` of that nested message which might be e.g. a `bool`.

As you traverse the input structures you must keep track of the path so that you can later look up the comment for it. This is all described in more detail in the [`SourceCodeInfo`](https://github.com/protocolbuffers/protobuf/blob/master/src/google/protobuf/descriptor.proto#L755) message. This results in more complex code :-(.

### Enums as Strings

Enums in protobuf's default Go output become a [custom `int32` type](https://developers.google.com/protocol-buffers/docs/reference/go-generated#enum) and a bunch of constant values, as well as maps to convert back and forth between their string representations. By default they marshal as integers. The `jsonpb` marshaller does some magic to make these strings in JSON.

For our plugin we want to always treat them as strings and to use Huma's default JSON & CBOR marshallers. This means defining our own `string` type and a constant for each value so that the enum is easy to use from service code. We use the protobuf-go generated maps to convert between the value and the `int32` when going from or to protobuf.

Overall this means a bit more generated code, but it provides a nicer interface in Go and guarantees strings in the marshalled output.

### Timestamps

[Timestamps](https://developers.google.com/protocol-buffers/docs/reference/google.protobuf#timestamp) are represented as normal Go `time.Time` instances to make them easier to work with. When going to protobuf, these get converted into `timestamppb.Timestamp` instances. When marshalled by Huma, the `time.Time` is represented as an ISO8601 string.

### Field Naming & Go Lint

While protobuf [got an exception](https://github.com/golang/go/wiki/CodeReviewComments#initialisms), all other code should capitalize initialisms and generally use camel casing in Go. We strive to be better and pass the linter. Therefore, each message, field, etc will have both a Huma name and a `ProtoGoName` that refers to the generated Go names to allow us to convert between the two. This makes the service code much more consistent and easier to maintain.

#### Nested Naming

Protobuf allows message and enum definitions to be nested. We output a flat structure where each name is the camel case variant of all the names combined, e.g. `.example.Foo.Bar.Baz` becomes `FooBarBaz` in the `example` package.

### Maps

Protobuf doesn't have maps. "But you can do `map<string, int32>`!" you try to object. That's just [syntactic sugar for backwards compatibility reasons](https://developers.google.com/protocol-buffers/docs/proto3#backwards_compatibility). In reality it generates a new message type with key/value fields and the original field becomes a `repeated` of that generated message type, which explains why you can't have repeated map fields. If multiple items in the array have the same key, the last one wins.

Needless to say that sucks to use in Go so we throw away the intermediate generated type and use Go `map` fields. This is the same behavior as the official Go protobuf plugin, it is mainly called out here because it unfortunately adds complexity to the implementation.

Note: map keys _must_ be strings for Huma to support them as the primary output format is JSON. CBOR does support non-string keys, so maybe that's something to consider for the future?

### Map & Array Assignment

Since maps and arrays of different value types can't be assigned to each other you will see loops in the generated code where we create the different map type and assign the converted value for each key. There's no way around this; it's just how Go works.

### One-of Support

This is an interesting one. Huma doesn't support one-of out of the box, despite [JSON-Schema having support for it](https://json-schema.org/draft/2019-09/json-schema-core.html#rfc.section.9.2.1). For now we expose individual fields. If you set multiple in the request JSON then you get a validation error.

Go generates an intermediate type and a wrapper struct for a single field. This is why our field representations have a `OneOf` attribute which corresponds to the single generated Go field name for all the possible fields in the one-of. This is used in the template to set the right field.

Using the official Go approach wouldn't work well for Huma as we can't easily reflect type information for all possible structs at runtime.

There is no such thing as a one-of on the wire. It's just plain fields each with their own field number. The one-of is [behavior when **setting** a field](https://developers.google.com/protocol-buffers/docs/proto3#oneof_features), which unsets the other fields in the group to ensure only a single field is transmitted. If for some reason multiple fields _are_ transmitted, the last one wins.

## Testing

There is an `example.proto` file that is used to exercise the features listed above in a Go test. Running the test itself is simple:

```sh
$ go generate && go test
```

Make sure to run `go generate` again any time you update the `example.proto` or update the protobuf compiler plugin code or update the `huma.proto` annotation definitions.
