package kitchensink_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/psyb0t/ctxerrors/commerr"
	"github.com/psyb0t/gen-asyncapi/asyncapirt"
	"github.com/psyb0t/gen-asyncapi/examples/kitchensink"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const colorRed = "red"

const validItem = `{
	"sku": "ABC-1234",
	"price": 9.99,
	"quantity": 3,
	"tags": ["red", "big"],
	"dimensions": {"width": 10, "height": 20},
	"attributes": {"x-color": "red"},
	"counters": {"views": 1}
}`

func TestServerURLs(t *testing.T) {
	t.Parallel()

	assert.Equal(
		t,
		"wss://eu.example.com:443/v2/ws",
		kitchensink.ServerProductionURL,
	)
	assert.Equal(
		t,
		"wss://us.example.com:8443/v1/ws",
		kitchensink.ServerProductionURLWith("us", "8443", "1"),
	)
	assert.Equal(t, "mqtt://broker.example.com", kitchensink.ServerBrokerURL)
	assert.Equal(
		t,
		"tenants/acme/commands",
		kitchensink.CommandsAddress("acme"),
	)
}

// withMember returns validItem with one member replaced or added.
func withMember(t *testing.T, name string, value any) string {
	t.Helper()

	var item map[string]any
	require.NoError(t, json.Unmarshal([]byte(validItem), &item))

	item[name] = value

	data, err := json.Marshal(item)
	require.NoError(t, err)

	return string(data)
}

func TestItemConstraints(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		input   string
		wantErr error
	}{
		{name: "valid", input: validItem},
		{
			name:    "sku pattern",
			input:   withMember(t, "sku", "abc"),
			wantErr: commerr.ErrValidationFailed,
		},
		{
			name:    "price not a multiple of 0.01",
			input:   withMember(t, "price", 9.999),
			wantErr: commerr.ErrValidationFailed,
		},
		{
			name:    "price not above the exclusive minimum",
			input:   withMember(t, "price", 0),
			wantErr: commerr.ErrValidationFailed,
		},
		{
			name:    "quantity above the maximum",
			input:   withMember(t, "quantity", 1001),
			wantErr: commerr.ErrValidationFailed,
		},
		{
			name:    "duplicate tags",
			input:   withMember(t, "tags", []string{colorRed, colorRed}),
			wantErr: commerr.ErrValidationFailed,
		},
		{
			name:    "tag shorter than minLength",
			input:   withMember(t, "tags", []string{"r"}),
			wantErr: commerr.ErrValidationFailed,
		},
		{
			name: "attribute key outside patternProperties",
			input: withMember(
				t,
				"attributes",
				map[string]string{"color": colorRed},
			),
			wantErr: commerr.ErrValidationFailed,
		},
		{
			name:    "counter key failing propertyNames",
			input:   withMember(t, "counters", map[string]int{"Views": 1}),
			wantErr: commerr.ErrValidationFailed,
		},
		{
			name: "too many counters",
			input: withMember(
				t,
				"counters",
				map[string]int{"a": 1, "b": 2, "c": 3, "d": 4},
			),
			wantErr: commerr.ErrValidationFailed,
		},
		{
			name:  "optional array with minItems present",
			input: withMember(t, "aliases", []string{colorRed}),
		},
		{
			name:    "optional array present but below minItems",
			input:   withMember(t, "aliases", []string{}),
			wantErr: commerr.ErrValidationFailed,
		},
		{
			name:    "short password",
			input:   withMember(t, "password", "short"),
			wantErr: commerr.ErrValidationFailed,
		},
		{
			name:    "unknown status",
			input:   withMember(t, "status", "lost"),
			wantErr: commerr.ErrValidationFailed,
		},
		{
			name:    "priority outside the integer enum",
			input:   withMember(t, "priority", 9),
			wantErr: commerr.ErrValidationFailed,
		},
		{
			name:    "ratio outside the number enum",
			input:   withMember(t, "ratio", 2.5),
			wantErr: commerr.ErrValidationFailed,
		},
		{
			name:    "undeclared property",
			input:   withMember(t, "bogus", 1),
			wantErr: commerr.ErrValidationFailed,
		},
		{
			name: "dimension width at the exclusive maximum",
			input: withMember(
				t,
				"dimensions",
				map[string]int{"width": 10000, "height": 1},
			),
			wantErr: commerr.ErrValidationFailed,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var item kitchensink.Item

			err := json.Unmarshal([]byte(tc.input), &item)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, "ABC-1234", item.Sku)
			assert.Equal(t, int32(20), item.Dimensions.Height)
			assert.Nil(t, item.Dimensions.Depth)
		})
	}
}

func TestItemPropertyCount(t *testing.T) {
	t.Parallel()

	var item kitchensink.Item

	tooFew := `{"sku":"ABC-1234","price":1,"quantity":1,"tags":["ab"]}`
	require.ErrorIs(
		t,
		json.Unmarshal([]byte(tooFew), &item),
		commerr.ErrRequiredFieldNotSet,
	)

	var dimensions kitchensink.Dimensions
	require.NoError(
		t,
		json.Unmarshal(
			[]byte(`{"width":1,"height":2,"depth":null}`),
			&dimensions,
		),
	)
	assert.Nil(t, dimensions.Depth)
}

func TestEventDiscriminatorFromAllOfConst(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		input     string
		wantValue any
		wantErr   error
	}{
		{
			name:      "deleted",
			input:     `{"kind":"deleted","id":"1","occurredAt":"2026-01-01T00:00:00Z","hard":true}`,
			wantValue: kitchensink.Deleted{},
		},
		{
			name:      "renamed",
			input:     `{"kind":"renamed","id":"1","occurredAt":"2026-01-01T00:00:00Z","from":"a","to":"b"}`,
			wantValue: kitchensink.Renamed{},
		},
		{
			name:    "renamed without the required from",
			input:   `{"kind":"renamed","id":"1","occurredAt":"2026-01-01T00:00:00Z","to":"b"}`,
			wantErr: commerr.ErrRequiredFieldNotSet,
		},
		{
			name:    "created without the required item",
			input:   `{"kind":"created","id":"1","occurredAt":"2026-01-01T00:00:00Z"}`,
			wantErr: commerr.ErrRequiredFieldNotSet,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var event kitchensink.Event

			err := json.Unmarshal([]byte(tc.input), &event)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)

			value, err := event.Value()
			require.NoError(t, err)
			assert.IsType(t, tc.wantValue, value)
		})
	}
}

func TestRenamedDependentRequired(t *testing.T) {
	t.Parallel()

	var renamed kitchensink.Renamed

	withoutFrom := `{"kind":"renamed","id":"1","occurredAt":"2026-01-01T00:00:00Z","to":"b"}`
	require.ErrorIs(
		t,
		json.Unmarshal([]byte(withoutFrom), &renamed),
		commerr.ErrRequiredFieldNotSet,
	)
}

func TestCommandPayloadAnyOf(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		input     string
		wantValue any
		wantErr   error
	}{
		{name: "string", input: `"reboot"`, wantValue: "reboot"},
		{name: "integer", input: `5`, wantValue: int64(5)},
		{
			name:      "dimensions",
			input:     `{"width":1,"height":2}`,
			wantValue: kitchensink.Dimensions{Width: 1, Height: 2},
		},
		{
			name:    "string over maxLength matches nothing",
			input:   `"` + strings.Repeat("x", 65) + `"`,
			wantErr: asyncapirt.ErrNoVariantMatched,
		},
		{
			name:    "nothing matches",
			input:   `[1]`,
			wantErr: asyncapirt.ErrNoVariantMatched,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var payload kitchensink.CommandPayload

			err := json.Unmarshal([]byte(tc.input), &payload)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)

			value, err := payload.Value()
			require.NoError(t, err)
			assert.Equal(t, tc.wantValue, value)
		})
	}
}

func TestHeadersAndTraitsAndCorrelation(t *testing.T) {
	t.Parallel()

	var headers kitchensink.TraceHeaders
	require.NoError(
		t,
		json.Unmarshal([]byte(`{"x-trace-id":"t","x-attempt":2}`), &headers),
	)
	require.ErrorIs(
		t,
		json.Unmarshal([]byte(`{"x-attempt":0}`), &headers),
		commerr.ErrValidationFailed,
	)

	assert.Equal(
		t,
		"event",
		kitchensink.MessageEvent,
		"the message name wins over the trait's name",
	)
	assert.Equal(
		t,
		"$message.payload#/commandId",
		kitchensink.MessageCommandCorrelationIDLocation,
	)
	assert.Equal(
		t,
		"$message.payload#/commandId",
		kitchensink.MessageCommandResultCorrelationIDLocation,
	)
}

func TestReceivedMessages(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    string
		wantName string
	}{
		{
			name:     "heartbeat const",
			input:    `{"beat":true}`,
			wantName: kitchensink.MessageHeartbeat,
		},
		{
			name:     "external ping from another file",
			input:    `{"ping":"external"}`,
			wantName: kitchensink.MessageExternalPing,
		},
		{
			name:     "inline payload",
			input:    `{"n":7}`,
			wantName: kitchensink.MessageTelemetryInlinePayload,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			message, err := kitchensink.DecodeReceivedMessage([]byte(tc.input))
			require.NoError(t, err)
			assert.Equal(t, tc.wantName, message.Name)
		})
	}

	// A heartbeat with the wrong const is not a heartbeat. Nothing typed
	// matches, so the Avro message, whose payload is opaque, takes it.
	message, err := kitchensink.DecodeReceivedMessage([]byte(`{"beat":false}`))
	require.NoError(t, err)
	assert.Equal(t, kitchensink.MessageAvroReading, message.Name)

	raw, ok := message.Payload.(json.RawMessage)
	require.True(t, ok)
	assert.JSONEq(t, `{"beat":false}`, string(raw))
}

func TestUnreferencedComponentsAreGenerated(t *testing.T) {
	t.Parallel()

	var tree kitchensink.Tree
	require.NoError(
		t,
		json.Unmarshal(
			[]byte(`{"value":"root","children":[{"value":"leaf"}]}`),
			&tree,
		),
	)
	require.Len(t, tree.Children, 1)

	var envelope kitchensink.Envelope
	require.NoError(
		t,
		json.Unmarshal(
			[]byte(
				`{"kind":"x","id":"1","occurredAt":"2026-01-01T00:00:00Z","traceId":null}`,
			),
			&envelope,
		),
	)
	assert.Nil(t, envelope.TraceID)

	assert.Equal(t, "TreeMessage", kitchensink.MessageTreeMessage)
}
