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
	"github.com/istreamlabs/protoc-gen-huma/example"
	"github.com/istreamlabs/protoc-gen-huma/examplehuma"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

//go:generate protoc --proto_path annotation annotation/huma.proto --go_out=./annotation --go_opt=paths=source_relative
//go:generate go install
//go:generate sh -c "mkdir -p example examplehuma && DUMP_REQUEST=1 protoc --proto_path=. -I=. --go_out=example --go_opt=paths=source_relative --huma_out=examplehuma example.proto"

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
	for k := range examplehuma.GlobalValuesMap {
		keys = append(keys, string(k))
	}

	assert.NotContains(t, keys, "TWO")
}

func TestHumaRoundtrip(t *testing.T) {
	// Example protobuf message we will use to test various features.
	proto := &example.Message{
		Hidden:     "hidden",
		Num32:      int32(2),
		Num64:      int64(2),
		Unsigned32: uint32(3),
		Unsigned64: uint64(4),
		Float:      float32(5.1),
		Double:     float64(6.1),
		Name:       "foo",
		Enabled:    true,
		Sub: &example.Sub{
			CamelCaseEnum: example.Sub_BAR,
			SnakeCaseEnum: 5, // Invalid value, will not serialize.
		},
		PrimitiveArray: []int32{1, 2, 3},
		EnumArray: []example.Global{
			example.Global_ONE,
			example.Global_TWO, // Note: this is NOT public and should not be included!
		},
		ComplexArray: []*example.Another{
			{Value: "first"},
			{Value: "second"},
		},
		Kv: map[string]int32{"a": 1, "b": 2},
		KvComplex: map[string]*example.Another{
			"complex": {Value: "map"},
		},
		OnlyOne: &example.Message_Another{
			Another: &example.Another{
				Value: "another",
			},
		},
		Ts:   timestamppb.New(time.Date(2020, 01, 01, 12, 0, 0, 0, time.UTC)),
		Mp2T: true,
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
		"mp2t": true
	}`

	// Set up a Huma instance & register a route. No middleware so that we
	// get stack traces if tests crash.
	app := huma.New("Test Router", "1.0.0")

	app.Resource("/").Get("get-message", "docs",
		responses.OK().Model([]examplehuma.Message{}),
	).Run(func(ctx huma.Context, input struct {
		Body examplehuma.Message
	}) {
		// Generate an external model from an internal proto input.
		gen := examplehuma.Message{}
		gen.FromProto(proto)

		// Round trip test taking in JSON, converting, and converting back.
		rt := examplehuma.Message{}
		rt.FromProto(input.Body.ToProto(nil))

		ctx.WriteModel(http.StatusOK, []examplehuma.Message{gen, rt})
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
