package shapes_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/psyb0t/ctxerrors"
	"github.com/psyb0t/ctxerrors/commerr"
	"github.com/psyb0t/gen-asyncapi/asyncapirt"
	"github.com/psyb0t/gen-asyncapi/examples/shapes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const validReading = `{"value":3.5,"unit":null,"tags":["a"],"level":-1,"ratio":0.5,"labels":{"x":1}}`

func TestMeasurementUnionWithoutDiscriminator(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		input     string
		wantValue any
		wantErr   error
	}{
		{name: "number variant", input: `3.5`, wantValue: 3.5},
		{
			name:      "string variant passes its pattern",
			input:     `"12kg"`,
			wantValue: "12kg",
		},
		{
			name:      "object variant",
			input:     `{"x":1,"y":2}`,
			wantValue: shapes.Point{X: 1, Y: 2},
		},
		{
			name:    "string failing its pattern matches nothing",
			input:   `"kg"`,
			wantErr: asyncapirt.ErrNoVariantMatched,
		},
		{
			name:    "object with unknown key matches nothing",
			input:   `{"x":1,"y":2,"z":3}`,
			wantErr: asyncapirt.ErrNoVariantMatched,
		},
		{
			name:    "bool matches nothing",
			input:   `true`,
			wantErr: asyncapirt.ErrNoVariantMatched,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var measurement shapes.Measurement

			err := json.Unmarshal([]byte(tc.input), &measurement)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)

			value, err := measurement.Value()
			require.NoError(t, err)
			assert.Equal(t, tc.wantValue, value)
		})
	}
}

func TestSensorReading(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		input   string
		wantErr error
	}{
		{name: "valid with null unit", input: validReading},
		{
			name:    "nullable required unit must still be present",
			input:   `{"value":1,"tags":["a"],"level":0,"ratio":0,"labels":{}}`,
			wantErr: commerr.ErrRequiredFieldNotSet,
		},
		{
			name:    "ratio at the exclusive maximum",
			input:   `{"value":1,"unit":"m","tags":["a"],"level":0,"ratio":1,"labels":{}}`,
			wantErr: commerr.ErrValidationFailed,
		},
		{
			name:    "label not above the exclusive minimum",
			input:   `{"value":1,"unit":"m","tags":["a"],"level":0,"ratio":0,"labels":{"x":0}}`,
			wantErr: commerr.ErrValidationFailed,
		},
		{
			name:    "empty tags below minItems",
			input:   `{"value":1,"unit":"m","tags":[],"level":0,"ratio":0,"labels":{}}`,
			wantErr: commerr.ErrValidationFailed,
		},
		{
			name:    "tag longer than maxLength",
			input:   `{"value":1,"unit":"m","tags":["123456789"],"level":0,"ratio":0,"labels":{}}`,
			wantErr: commerr.ErrValidationFailed,
		},
		{
			name:    "level outside the enum",
			input:   `{"value":1,"unit":"m","tags":["a"],"level":7,"ratio":0,"labels":{}}`,
			wantErr: commerr.ErrValidationFailed,
		},
		{
			name:    "label overflows int32",
			input:   `{"value":1,"unit":"m","tags":["a"],"level":0,"ratio":0,"labels":{"x":4294967296}}`,
			wantErr: commerr.ErrCouldNotDecode,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var reading shapes.SensorReading

			err := json.Unmarshal([]byte(tc.input), &reading)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
			assert.Nil(t, reading.Unit)
			assert.Equal(t, shapes.LevelMinus1, reading.Level)
			assert.InDelta(t, 0.5, reading.Ratio, 0.0001)
			assert.Equal(t, map[string]int32{"x": 1}, reading.Labels)
		})
	}
}

func TestSensorReadingMarshalKeepsNullableAndDropsOptional(t *testing.T) {
	t.Parallel()

	var value shapes.Measurement
	require.NoError(t, value.FromFloat64(1))

	data, err := json.Marshal(shapes.SensorReading{
		Value:  value,
		Tags:   shapes.Tags{"a"},
		Labels: map[string]int32{},
	})
	require.NoError(t, err)

	assert.JSONEq(
		t,
		`{"value":1,"unit":null,"tags":["a"],"level":0,"ratio":0,"labels":{}}`,
		string(data),
	)
}

func TestRecursiveNode(t *testing.T) {
	t.Parallel()

	var node shapes.Node
	require.NoError(
		t,
		json.Unmarshal(
			[]byte(
				`{"name":"a","children":[{"name":"b","children":[{"name":"c"}]}]}`,
			),
			&node,
		),
	)
	require.Len(t, node.Children, 1)
	require.Len(t, node.Children[0].Children, 1)
	assert.Equal(t, "c", node.Children[0].Children[0].Name)

	err := json.Unmarshal(
		[]byte(`{"name":"a","children":[{"children":[]}]}`),
		&node,
	)
	require.ErrorIs(t, err, commerr.ErrRequiredFieldNotSet)
}

func TestFrameImplicitDiscriminator(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		input     string
		wantValue any
		wantErr   error
	}{
		{
			name:      "const type selects StartedFrame",
			input:     `{"type":"run.started","sequence":1}`,
			wantValue: shapes.StartedFrame{Type: "run.started", Sequence: 1},
		},
		{
			name:  "const type selects StoppedFrame",
			input: `{"type":"run.stopped","sequence":2,"reason":"done"}`,
			wantValue: shapes.StoppedFrame{
				Type:     "run.stopped",
				Sequence: 2,
				Reason:   new("done"),
			},
		},
		{
			name:    "unknown type",
			input:   `{"type":"run.paused","sequence":3}`,
			wantErr: asyncapirt.ErrUnknownDiscriminator,
		},
		{
			name:    "selected variant missing a field",
			input:   `{"type":"run.started"}`,
			wantErr: commerr.ErrRequiredFieldNotSet,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			message, err := shapes.DecodeReceivedMessage([]byte(tc.input))
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
			require.Equal(t, shapes.MessageFrame, message.Name)

			frame, ok := message.Payload.(shapes.FramePayload)
			require.True(t, ok)

			value, err := frame.Value()
			require.NoError(t, err)
			assert.Equal(t, tc.wantValue, value)
		})
	}

	var frame shapes.FramePayload
	require.NoError(t, frame.FromStoppedFrame(shapes.StoppedFrame{Sequence: 4}))

	data, err := json.Marshal(frame)
	require.NoError(t, err)
	assert.JSONEq(t, `{"type":"run.stopped","sequence":4}`, string(data))
}

func TestBatchArrayAndMapVariants(t *testing.T) {
	t.Parallel()

	var batch shapes.Batch
	require.NoError(t, json.Unmarshal([]byte(`["a","b"]`), &batch))

	value, err := batch.Value()
	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b"}, value)

	require.NoError(t, json.Unmarshal([]byte(`{"a":1}`), &batch))

	asMap, err := batch.AsMapOfInt64()
	require.NoError(t, err)
	assert.Equal(t, map[string]int64{"a": 1}, asMap)

	require.ErrorIs(
		t,
		json.Unmarshal([]byte(`"x"`), &batch),
		asyncapirt.ErrNoVariantMatched,
	)

	require.NoError(t, batch.FromArrayOfString([]string{"z"}))
	require.NoError(t, batch.Validate())
}

func TestMeasurementFromVariants(t *testing.T) {
	t.Parallel()

	var measurement shapes.Measurement

	require.NoError(t, measurement.FromString("5kg"))
	require.NoError(t, measurement.Validate())

	require.NoError(t, measurement.FromPoint(shapes.Point{X: 1, Y: 2}))

	data, err := json.Marshal(measurement)
	require.NoError(t, err)
	assert.JSONEq(t, `{"x":1,"y":2}`, string(data))

	require.NoError(t, measurement.FromString("not a measurement"))
	require.ErrorIs(t, measurement.Validate(), asyncapirt.ErrNoVariantMatched)
}

type frameHandler struct {
	shapes.UnimplementedHandler

	frames chan shapes.FramePayload
}

func (h frameHandler) HandleFrame(
	_ context.Context,
	payload shapes.FramePayload,
) error {
	h.frames <- payload

	return nil
}

func TestRunDispatchesFrames(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	transport := &queueTransport{incoming: make(chan []byte, 4)}
	handler := frameHandler{frames: make(chan shapes.FramePayload, 1)}

	runErr := make(chan error, 1)

	go func() {
		runErr <- shapes.NewClient(transport).Run(ctx, handler)
	}()

	transport.incoming <- []byte(`{"name":"ignored by UnimplementedHandler"}`)

	transport.incoming <- []byte(`not json`)

	transport.incoming <- []byte(`{"type":"run.started","sequence":1}`)

	frame := <-handler.frames
	discriminator, err := frame.Discriminator()
	require.NoError(t, err)
	assert.Equal(t, "run.started", discriminator)

	cancel()
	require.ErrorIs(t, <-runErr, context.Canceled)

	failing := &queueTransport{
		incoming: make(chan []byte),
		readErr:  commerr.ErrReadFailed,
	}
	require.ErrorIs(
		t,
		shapes.NewClient(failing).Run(context.Background(), handler),
		commerr.ErrReadFailed,
	)
	require.NoError(t, shapes.NewClient(failing).Close())
}

type queueTransport struct {
	incoming chan []byte
	readErr  error
}

func (q *queueTransport) WriteMessage(
	context.Context,
	[]byte,
) error {
	return nil
}

func (q *queueTransport) ReadMessage(ctx context.Context) ([]byte, error) {
	if q.readErr != nil {
		return nil, q.readErr
	}

	select {
	case data := <-q.incoming:
		return data, nil
	case <-ctx.Done():
		return nil, ctxerrors.Wrap(ctx.Err(), "read message")
	}
}

func (q *queueTransport) Close() error { return nil }

func TestOpenAPIDiscriminatorMapping(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		input     string
		wantValue any
		wantErr   error
	}{
		{
			name:      "ref mapping",
			input:     `{"kind":"alpha","alpha":1}`,
			wantValue: shapes.SensorA{Kind: "alpha", Alpha: 1},
		},
		{
			name:      "bare name mapping",
			input:     `{"kind":"beta","beta":"b"}`,
			wantValue: shapes.SensorB{Kind: "beta", Beta: "b"},
		},
		{
			name:      "unmapped variant answers to its schema name",
			input:     `{"kind":"SensorC","gamma":true}`,
			wantValue: shapes.SensorC{Kind: "SensorC", Gamma: true},
		},
		{
			name:    "a mapped schema's name is not a value",
			input:   `{"kind":"SensorA","alpha":1}`,
			wantErr: asyncapirt.ErrUnknownDiscriminator,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var source shapes.SensorReadingSource

			err := json.Unmarshal([]byte(tc.input), &source)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)

			value, err := source.Value()
			require.NoError(t, err)
			assert.Equal(t, tc.wantValue, value)
		})
	}
}

func TestFeedAddressKeywordParameter(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "feed/news/u1", shapes.FeedAddress("news", "u1"))
}

func TestReceiveWithoutPinnedMessages(t *testing.T) {
	t.Parallel()

	message, err := shapes.DecodeReceivedMessage([]byte(validReading))
	require.NoError(t, err)
	assert.Equal(t, shapes.MessageReading, message.Name)

	message, err = shapes.DecodeReceivedMessage([]byte(`{"name":"root"}`))
	require.NoError(t, err)
	assert.Equal(t, shapes.MessageTree, message.Name)
}

type recordingTransport struct {
	written [][]byte
}

func (r *recordingTransport) WriteMessage(
	_ context.Context,
	data []byte,
) error {
	r.written = append(r.written, data)

	return nil
}

func (r *recordingTransport) ReadMessage(ctx context.Context) ([]byte, error) {
	<-ctx.Done()

	return nil, ctxerrors.Wrap(ctx.Err(), "read message")
}

func (r *recordingTransport) Close() error { return nil }

func TestPublishTree(t *testing.T) {
	t.Parallel()

	transport := &recordingTransport{}
	client := shapes.NewClient(transport)

	require.NoError(
		t,
		client.PublishTree(context.Background(), shapes.Node{Name: "root"}),
	)
	require.Len(t, transport.written, 1)
	assert.JSONEq(t, `{"name":"root"}`, string(transport.written[0]))
}
