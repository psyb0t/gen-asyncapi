package spec_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/psyb0t/ctxerrors/commerr"
	"github.com/psyb0t/gogen-asyncapi/internal/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScalarTypes(t *testing.T) {
	t.Parallel()

	document, err := spec.NewResolver().Document(
		filepath.Join("testdata", "scalars.yaml"),
	)
	require.NoError(t, err)

	root, ok := document.(*spec.Object)
	require.True(t, ok)

	testCases := []struct {
		key  string
		want any
	}{
		{key: "int", want: int64(42)},
		{key: "hex", want: int64(16)},
		{key: "float", want: 1.5},
		{key: "bool", want: true},
		{key: "null_value", want: nil},
		{key: "text", want: "hello"},
		{key: "quoted_number", want: "7"},
		{key: "tagged_float", want: 3.0},
		{key: "list", want: []any{int64(1), "two"}},
	}

	for _, tc := range testCases {
		t.Run(tc.key, func(t *testing.T) {
			t.Parallel()

			value, exists := root.Get(tc.key)
			require.True(t, exists)
			assert.Equal(t, tc.want, value)
		})
	}

	alias := root.Object("alias")
	require.NotNil(t, alias)
	assert.Equal(t, int64(1), alias.Values["a"])

	assert.Empty(t, root.String("int"), "String returns empty for non-strings")
	assert.Nil(t, root.Object("text"), "Object returns nil for non-mappings")

	var missing *spec.Object

	_, exists := missing.Get("x")
	assert.False(t, exists)
}

func TestScalarParseErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		content string
	}{
		{name: "bad int", content: "x: !!int nope\n"},
		{name: "bad float", content: "x: !!float nope\n"},
		{name: "bad bool", content: "x: !!bool nope\n"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "bad.yaml")
			require.NoError(t, os.WriteFile(path, []byte(tc.content), 0o600))

			_, err := spec.NewResolver().Document(path)
			require.Error(t, err)
		})
	}

	empty := filepath.Join(t.TempDir(), "empty.yaml")
	require.NoError(t, os.WriteFile(empty, []byte(""), 0o600))

	_, err := spec.Load(empty)
	require.ErrorIs(t, err, commerr.ErrParseFailed)
}
