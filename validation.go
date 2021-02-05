package main

import (
	"strings"

	"github.com/envoyproxy/protoc-gen-validate/validate"
	"github.com/istreamlabs/protoc-gen-huma/annotation"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

// Validation represents Huma-supported validation rules. These follow JSON
// Schema closely.
type Validation struct {
	ReadOnly         bool
	Deprecated       bool
	IsRequired       bool
	Minimum          float64
	ExclusiveMinimum float64
	Maximum          float64
	ExclusiveMaximum float64
	MinLength        int64
	MaxLengh         int64
	Pattern          string
	Format           string
	MinItems         int64
	MaxItems         int64
	Unique           bool
	EnumValues       []string
}

// convertValidation from protoc-gen-validate rules to Huma rules.
func convertValidation(protoField *descriptorpb.FieldDescriptorProto, f *Field) {
	if proto.GetExtension(protoField.GetOptions(), annotation.E_ReadOnly).(bool) {
		f.Validation.ReadOnly = true
	}

	if protoField.Options != nil && protoField.Options.Deprecated != nil && *protoField.Options.Deprecated {
		f.Validation.Deprecated = true
		f.Comment = strings.TrimSpace("Deprecated: Do not use. " + f.Comment)
	}

	// Default enum settings include all values. This is on each field because
	// subsequent validation rules can disable certain enum values for *just*
	// this particular field.
	if f.Enum != nil {
		values := []string{}
		for _, v := range f.Enum.Values {
			values = append(values, v.Label)
		}
		f.Validation.EnumValues = values
	}

	if rules, ok := proto.GetExtension(protoField.GetOptions(), validate.E_Rules).(*validate.FieldRules); ok && rules != nil {
		// Message rules, e.g. required fields.
		if rules.Message != nil && rules.Message.Required != nil && *rules.Message.Required {
			f.Validation.IsRequired = true
		}

		// Number rules. Each type gets its own struct... ugh. We just support the
		// most common ones.
		// TODO: add more types?
		if i := rules.GetInt32(); i != nil {
			if i.Gte != nil {
				f.Validation.Minimum = float64(*i.Gte)
			}
			if i.Gt != nil {
				f.Validation.ExclusiveMinimum = float64(*i.Gt)
			}
			if i.Lte != nil {
				f.Validation.Minimum = float64(*i.Lte)
			}
			if i.Lt != nil {
				f.Validation.ExclusiveMaximum = float64(*i.Gt)
			}
		}

		if i := rules.GetInt64(); i != nil {
			if i.Gte != nil {
				f.Validation.Minimum = float64(*i.Gte)
			}
			if i.Gt != nil {
				f.Validation.ExclusiveMinimum = float64(*i.Gt)
			}
			if i.Lte != nil {
				f.Validation.Minimum = float64(*i.Lte)
			}
			if i.Lt != nil {
				f.Validation.ExclusiveMaximum = float64(*i.Gt)
			}
		}

		if i := rules.GetUint32(); i != nil {
			if i.Gte != nil {
				f.Validation.Minimum = float64(*i.Gte)
			}
			if i.Gt != nil {
				f.Validation.ExclusiveMinimum = float64(*i.Gt)
			}
			if i.Lte != nil {
				f.Validation.Minimum = float64(*i.Lte)
			}
			if i.Lt != nil {
				f.Validation.ExclusiveMaximum = float64(*i.Gt)
			}
		}

		if i := rules.GetUint64(); i != nil {
			if i.Gte != nil {
				f.Validation.Minimum = float64(*i.Gte)
			}
			if i.Gt != nil {
				f.Validation.ExclusiveMinimum = float64(*i.Gt)
			}
			if i.Lte != nil {
				f.Validation.Minimum = float64(*i.Lte)
			}
			if i.Lt != nil {
				f.Validation.ExclusiveMaximum = float64(*i.Gt)
			}
		}

		if i := rules.GetFloat(); i != nil {
			if i.Gte != nil {
				f.Validation.Minimum = float64(*i.Gte)
			}
			if i.Gt != nil {
				f.Validation.ExclusiveMinimum = float64(*i.Gt)
			}
			if i.Lte != nil {
				f.Validation.Minimum = float64(*i.Lte)
			}
			if i.Lt != nil {
				f.Validation.ExclusiveMaximum = float64(*i.Gt)
			}
		}

		if i := rules.GetDouble(); i != nil {
			if i.Gte != nil {
				f.Validation.Minimum = float64(*i.Gte)
			}
			if i.Gt != nil {
				f.Validation.ExclusiveMinimum = float64(*i.Gt)
			}
			if i.Lte != nil {
				f.Validation.Minimum = float64(*i.Lte)
			}
			if i.Lt != nil {
				f.Validation.ExclusiveMaximum = float64(*i.Gt)
			}
		}

		// String rules, e.g. min/max length and regular expression patterns.
		if s := rules.GetString_(); s != nil {
			if s.MinLen != nil {
				f.Validation.MinLength = int64(*s.MinLen)
			}
			if s.MaxLen != nil {
				f.Validation.MaxLengh = int64(*s.MaxLen)
			}
			if s.Pattern != nil {
				f.Validation.Pattern = *s.Pattern
			}
			if s.GetUri() {
				f.Validation.Format = "uri"
			}
			if s.GetUriRef() {
				f.Validation.Format = "uri-reference"
			}
		}

		// Array rules, e.g. min/max number of items.
		if r := rules.GetRepeated(); r != nil {
			if r.MinItems != nil {
				f.Validation.MinItems = int64(*r.MinItems)
			}

			if r.MaxItems != nil {
				f.Validation.MaxItems = int64(*r.MaxItems)
			}

			if r.Unique != nil {
				f.Validation.Unique = bool(*r.Unique)
			}
		}

		// Enum rules, e.g. filtering allowed values.
		if f.Enum != nil {
			values := []string{}
			notIn := []int32{}
			if e := rules.GetEnum(); e != nil {
				if e.NotIn != nil {
					notIn = e.NotIn
				}
			}

		outer:
			for _, v := range f.Enum.Values {
				for _, num := range notIn {
					if num == v.Value {
						continue outer
					}
				}
				values = append(values, v.Label)
			}
			f.Validation.EnumValues = values
		}
	}
}
