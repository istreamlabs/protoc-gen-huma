package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/istreamlabs/huma"
	"github.com/istreamlabs/huma/responses"
	"github.com/istreamlabs/protoc-gen-huma/example/package1"
	"github.com/istreamlabs/protoc-gen-huma/example/package1huma"
	"github.com/istreamlabs/protoc-gen-huma/example/package2"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

//go:generate protoc --proto_path annotation annotation/huma.proto --go_out=./annotation --go_opt=paths=source_relative
//go:generate go install
//go:generate sh -c "rm -rf example && mkdir -p example && DUMP_REQUEST=1 protoc --proto_path=./proto -I=. --go_out=example --go_opt=paths=source_relative --huma_out=example proto/package1/* proto/package2/*"

func TestMain(m *testing.M) {
	// Run the code generator to get proper coverage reporting. We don't care
	// about the output, just want to exercise the various code paths and be
	// able to see what was hit or missed. It's kind of hacky... but it works!
	input, _ := ioutil.ReadFile("request.pb")
	run(input)

	os.Exit(m.Run())
}

func TestExcludedEnum(t *testing.T) {
	keys := []string{}
	for k := range package1huma.GlobalValuesMap {
		keys = append(keys, string(k))
	}

	assert.NotContains(t, keys, "TWO")
}

func TestHumaRoundtrip(t *testing.T) {
	// Example protobuf message we will use to test various features.
	proto := &package1.Message{
		Hidden:     "hidden",
		Num32:      int32(2),
		Num64:      int64(2),
		Unsigned32: uint32(3),
		Unsigned64: uint64(4),
		Float:      float32(5.1),
		Double:     float64(6.1),
		Name:       "foo",
		Enabled:    true,
		Sub: &package1.Sub{
			CamelCaseEnum: package1.Sub_BAR,
			SnakeCaseEnum: 5, // Invalid value, will not serialize.
		},
		PrimitiveArray: []int32{1, 2, 3},
		EnumArray: []package1.Global{
			package1.Global_ONE,
			package1.Global_TWO, // Note: this is NOT public and should not be included!
		},
		ComplexArray: []*package1.Another{
			{Value: "first"},
			{Value: "second"},
		},
		Kv: map[string]int32{"a": 1, "b": 2},
		KvComplex: map[string]*package1.Another{
			"complex": {Value: "map"},
		},
		OnlyOne: &package1.Message_Another{
			Another: &package1.Another{
				Value: "another",
			},
		},
		Ts:   timestamppb.New(time.Date(2020, 01, 01, 12, 0, 0, 0, time.UTC)),
		Mp2T: true,
		CrossPackage: &package2.Message{
			Name: "crosspkg",
		},
	}

	// Expected JSON representation of the above. We will use this to both check
	// the above converted to JSON *and* do a round-trip test.
	json := `{
		"num32": 2,
		"num64": 2,
		"unsigned32": 3,
		"unsigned64": 4,
		"float": 5.1,
		"double": 6.1,
		"name": "foo",
		"enabled": true,
		"sub": {
			"camel_case_enum": "BAR"
		},
		"primitive_array": [1, 2, 3],
		"enum_array": ["ONE"],
		"complex_array": [
			{"value": "first"},
			{"value": "second"}
		],
		"kv": {
			"a": 1,
			"b": 2
		},
		"kv_complex": {
			"complex": {"value": "map"}
		},
		"another": {
			"value": "another"
		},
		"ts": "2020-01-01T12:00:00Z",
		"mp2t": true,
		"cross_package": {
			"name": "crosspkg"
		}
	}`

	// Set up a Huma instance & register a route. No middleware so that we
	// get stack traces if tests crash.
	app := huma.New("Test Router", "1.0.0")

	app.Resource("/").Get("get-message", "docs",
		responses.OK().Model([]package1huma.Message{}),
	).Run(func(ctx huma.Context, input struct {
		Body package1huma.Message
	}) {
		// Generate an external model from an internal proto input.
		gen := package1huma.Message{}
		gen.FromProto(proto)

		// Round trip test taking in JSON, converting, and converting back.
		rt := package1huma.Message{}
		rt.FromProto(input.Body.ToProto(nil))

		ctx.WriteModel(http.StatusOK, []package1huma.Message{gen, rt})
	})

	// Make a request against the service.
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", strings.NewReader(json))
	app.ServeHTTP(w, req)

	// Assert the response is as expected.
	assert.Equal(t, http.StatusOK, w.Code)

	// Note the `hidden` field was not made public so it not included below!
	assert.JSONEq(t, "["+json+","+json+"]", w.Body.String())

	// Test that validation is working as expected
	// First: exclusive minimum should fail
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/", strings.NewReader(`{"num32": 0}`))
	app.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code) // Fails because not > 0

	// Next: multiple of validation
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/", strings.NewReader(`{"num32": 1}`))
	app.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code) // Fails because not multiple of 2

	// Next: inclusive minimum should pass at the boundary value
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/", strings.NewReader(`{"num64": 0}`))
	app.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code) // Works because >= 0

	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/", strings.NewReader(`{"num64": -1}`))
	app.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code) // Fails because not >= 0

	// One-of test
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/", strings.NewReader(`{"tag": "foo", "another": {"value": "foo"}}`))
	app.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code) // Fails because of one-of rule
	assert.Contains(t, w.Body.String(), "tag, another")
}
