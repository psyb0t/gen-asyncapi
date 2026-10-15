package asyncapirt_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/psyb0t/ctxerrors"
	"github.com/psyb0t/ctxerrors/commerr"
	"github.com/psyb0t/gen-asyncapi/asyncapirt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	strictName     = "strict"
	kindBInput     = `{"kind":"b"}`
	quotedX        = `"x"`
	escapedPointer = "/a~1b/~0c/1"
)

type strictKind struct {
	Kind string `json:"kind"`
}

func (s *strictKind) UnmarshalJSON(data []byte) error {
	members, err := asyncapirt.DecodeObject(data, "strictKind")
	if err != nil {
		return ctxerrors.Wrap(err, "decode object")
	}

	if err := asyncapirt.RequireKeys(
		"strictKind",
		members,
		"kind",
	); err != nil {
		return ctxerrors.Wrap(err, "require keys")
	}

	if err := asyncapirt.RejectUnknownKeys(
		"strictKind",
		members,
		"kind",
	); err != nil {
		return ctxerrors.Wrap(err, "reject unknown keys")
	}

	type plain strictKind

	var decoded plain
	if err := json.Unmarshal(data, &decoded); err != nil {
		return ctxerrors.Wrap(
			asyncapirt.DecodeError("strictKind", err),
			"decode plain",
		)
	}

	*s = strictKind(decoded)

	return nil
}

func decodeInto[T any](data []byte) (any, error) {
	var value T
	if err := asyncapirt.UnmarshalStrict(data, &value); err != nil {
		return nil, ctxerrors.Wrap(err, "unmarshal strict")
	}

	return value, nil
}

func TestDecodeMessage(t *testing.T) {
	t.Parallel()

	anyObject := asyncapirt.Candidate{
		Name:   "anyObject",
		Decode: decodeInto[map[string]any],
	}
	strict := asyncapirt.Candidate{
		Name:   strictName,
		Decode: decodeInto[strictKind],
	}
	pinnedA := asyncapirt.Candidate{
		Name:     "pinnedA",
		Matchers: []asyncapirt.Matcher{{Pointer: "/kind", Value: `"a"`}},
		Decode:   decodeInto[strictKind],
	}
	fallback := asyncapirt.Candidate{
		Name:     "fallback",
		Decode:   decodeInto[json.RawMessage],
		Fallback: true,
	}
	pinnedNumber := asyncapirt.Candidate{
		Name:     "pinnedNumber",
		Matchers: []asyncapirt.Matcher{{Pointer: "/items/1/n", Value: `2`}},
		Decode:   decodeInto[map[string]any],
	}

	testCases := []struct {
		name       string
		input      string
		candidates []asyncapirt.Candidate
		wantName   string
		wantErr    error
	}{
		{
			name:       "pinned candidate wins over structural ones",
			input:      `{"kind":"a"}`,
			candidates: []asyncapirt.Candidate{anyObject, pinnedA},
			wantName:   "pinnedA",
		},
		{
			name:       "failed matcher excludes the candidate",
			input:      kindBInput,
			candidates: []asyncapirt.Candidate{pinnedA, strict},
			wantName:   strictName,
		},
		{
			name:       "numbers in arrays compare by value",
			input:      `{"items":[{"n":1},{"n":2.0}]}`,
			candidates: []asyncapirt.Candidate{strict, pinnedNumber},
			wantName:   "pinnedNumber",
		},
		{
			name:       "typed candidate beats a fallback",
			input:      kindBInput,
			candidates: []asyncapirt.Candidate{fallback, strict},
			wantName:   strictName,
		},
		{
			name:       "fallback takes what nothing typed matches",
			input:      `{"other":1}`,
			candidates: []asyncapirt.Candidate{strict, fallback},
			wantName:   "fallback",
		},
		{
			name:       "two structural matches are ambiguous",
			input:      kindBInput,
			candidates: []asyncapirt.Candidate{strict, anyObject},
			wantErr:    asyncapirt.ErrAmbiguousMessage,
		},
		{
			name:       "nothing decodes",
			input:      `{"other":1}`,
			candidates: []asyncapirt.Candidate{strict},
			wantErr:    asyncapirt.ErrUnknownMessage,
		},
		{
			name:       "no candidates",
			input:      `{}`,
			candidates: nil,
			wantErr:    asyncapirt.ErrUnknownMessage,
		},
		{
			name:       "malformed json",
			input:      `{"kind":`,
			candidates: []asyncapirt.Candidate{strict},
			wantErr:    commerr.ErrCouldNotDecode,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			name, payload, err := asyncapirt.DecodeMessage(
				[]byte(tc.input),
				tc.candidates,
			)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.wantName, name)
			assert.NotNil(t, payload)
		})
	}
}

func TestPointerEquals(t *testing.T) {
	t.Parallel()

	var document any
	require.NoError(
		t,
		json.Unmarshal([]byte(`{"a/b":{"~c":[true,"x",3]}}`), &document),
	)

	testCases := []struct {
		name     string
		pointer  string
		expected string
		want     bool
	}{
		{
			name:     "escaped tokens",
			pointer:  escapedPointer,
			expected: quotedX,
			want:     true,
		},
		{name: "bool", pointer: "/a~1b/~0c/0", expected: `true`, want: true},
		{name: "number", pointer: "/a~1b/~0c/2", expected: `3`, want: true},
		{
			name:     "different value",
			pointer:  escapedPointer,
			expected: `"y"`,
			want:     false,
		},
		{name: "missing key", pointer: "/nope", expected: quotedX, want: false},
		{
			name:     "index out of range",
			pointer:  "/a~1b/~0c/9",
			expected: quotedX,
			want:     false,
		},
		{
			name:     "step into scalar",
			pointer:  "/a~1b/~0c/1/x",
			expected: quotedX,
			want:     false,
		},
		{
			name:     "bad expected json",
			pointer:  escapedPointer,
			expected: `x`,
			want:     false,
		},
		{
			name:     "root",
			pointer:  "",
			expected: `{"a/b":{"~c":[true,"x",3]}}`,
			want:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(
				t,
				tc.want,
				asyncapirt.PointerEquals(document, tc.pointer, tc.expected),
			)
		})
	}
}

func TestDiscriminatorValue(t *testing.T) {
	t.Parallel()

	value, err := asyncapirt.DiscriminatorValue(
		[]byte(`{"type":"x"}`),
		"U",
		"type",
	)
	require.NoError(t, err)
	assert.Equal(t, "x", value)

	value, err = asyncapirt.DiscriminatorValue(
		[]byte(`{"type": 4}`),
		"U",
		"type",
	)
	require.NoError(t, err)
	assert.Equal(t, "4", value)

	_, err = asyncapirt.DiscriminatorValue([]byte(`{}`), "U", "type")
	require.ErrorIs(t, err, commerr.ErrRequiredFieldNotSet)

	_, err = asyncapirt.DiscriminatorValue([]byte(`null`), "U", "type")
	require.ErrorIs(t, err, commerr.ErrCouldNotDecode)
}

func TestSingleMatch(t *testing.T) {
	t.Parallel()

	value, err := asyncapirt.SingleMatch("U", []any{1})
	require.NoError(t, err)
	assert.Equal(t, 1, value)

	_, err = asyncapirt.SingleMatch("U", nil)
	require.ErrorIs(t, err, asyncapirt.ErrNoVariantMatched)

	_, err = asyncapirt.SingleMatch("U", []any{1, 2})
	require.ErrorIs(t, err, asyncapirt.ErrAmbiguousVariant)
}

func TestDecodeErrorKeepsNestedChain(t *testing.T) {
	t.Parallel()

	var wrapper struct {
		Inner strictKind `json:"inner"`
	}

	err := json.Unmarshal([]byte(`{"inner":{"kind":"a","extra":1}}`), &wrapper)
	require.ErrorIs(
		t,
		asyncapirt.DecodeError("wrapper", err),
		commerr.ErrValidationFailed,
	)

	err = json.Unmarshal([]byte(`{"inner":5}`), &wrapper)
	require.ErrorIs(
		t,
		asyncapirt.DecodeError("wrapper", err),
		commerr.ErrCouldNotDecode,
	)
}

type failingTransport struct {
	err     error
	written [][]byte
}

func (f *failingTransport) WriteMessage(_ context.Context, data []byte) error {
	f.written = append(f.written, data)

	return f.err
}

func (f *failingTransport) ReadMessage(context.Context) ([]byte, error) {
	return nil, f.err
}

func (f *failingTransport) Close() error { return f.err }

type alwaysInvalid struct{}

func (alwaysInvalid) Validate() error {
	return ctxerrors.Wrap(
		asyncapirt.Invalid("alwaysInvalid", "never valid"),
		"always invalid",
	)
}

func TestSend(t *testing.T) {
	t.Parallel()

	ok := &failingTransport{}
	require.NoError(
		t,
		asyncapirt.Send(context.Background(), ok, map[string]int{"a": 1}),
	)
	require.Len(t, ok.written, 1)
	assert.JSONEq(t, `{"a":1}`, string(ok.written[0]))

	invalid := &failingTransport{}
	require.ErrorIs(
		t,
		asyncapirt.Send(context.Background(), invalid, alwaysInvalid{}),
		commerr.ErrValidationFailed,
	)
	assert.Empty(t, invalid.written)

	broken := &failingTransport{err: commerr.ErrWriteFailed}
	require.ErrorIs(
		t,
		asyncapirt.Send(context.Background(), broken, 1),
		commerr.ErrWriteFailed,
	)

	require.ErrorIs(
		t,
		asyncapirt.Send(context.Background(), ok, make(chan int)),
		commerr.ErrInvalidValue,
	)
}
