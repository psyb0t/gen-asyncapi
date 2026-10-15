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

func TestLoad(t *testing.T) {
	t.Parallel()

	document, err := spec.Load(filepath.Join("testdata", "root.yaml"))
	require.NoError(t, err)

	assert.Equal(t, "3.0.1", document.AsyncAPIVersion)
	assert.Equal(t, "Root", document.Title)
	assert.Equal(t, "2.0.0", document.Version)

	require.Len(t, document.Servers, 1)
	assert.Equal(t, "wss://example.com/socket", document.Servers[0].URL())

	require.Len(t, document.Channels, 2)
	assert.Equal(
		t,
		"zeta",
		document.Channels[0].Name,
		"channels keep document order",
	)
	assert.Equal(t, "alpha", document.Channels[1].Name)

	alpha := document.Channels[1]
	require.Len(t, alpha.Parameters, 1)
	assert.Equal(t, []string{"one", "two"}, alpha.Parameters[0].Enum)
	assert.Equal(t, "one", alpha.Parameters[0].Default)

	require.Len(
		t,
		document.Messages,
		2,
		"a message shared across files and channels is one message",
	)
	shared := document.Messages[0]
	assert.Equal(t, "Shared", shared.Key)
	assert.Equal(t, "shared", shared.Name)
	assert.Equal(t, "/components/schemas/Payload", shared.PayloadAt.Pointer)
	assert.Equal(t, "shared.yaml", filepath.Base(shared.PayloadAt.File))
	assert.Same(t, shared, alpha.Messages[0])

	inline := document.Messages[1]
	assert.Equal(t, "zeta_inline", inline.Key)

	payload, ok := inline.Payload.(*spec.Object)
	require.True(t, ok)
	assert.Equal(
		t,
		[]string{"b", "a"},
		payload.Object("properties").Keys,
		"properties keep document order",
	)

	require.Len(t, document.Operations, 2)
	listen := document.Operations[0]
	assert.Equal(t, spec.ActionReceive, listen.Action)
	assert.Equal(
		t,
		document.Channels[0].Messages,
		listen.Messages,
		"no messages list means every channel message",
	)

	talk := document.Operations[1]
	assert.Equal(t, []*spec.Message{shared}, talk.Messages)
	require.NotNil(t, talk.Reply)
	assert.Same(t, document.Channels[0], talk.Reply.Channel)
	assert.Equal(t, []*spec.Message{inline}, talk.Reply.Messages)
}

func TestLoadErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		file    string
		wantErr error
	}{
		{
			name:    "asyncapi 2 is rejected",
			file:    "v2.yaml",
			wantErr: commerr.ErrInvalidValue,
		},
		{
			name:    "ref cycle",
			file:    "cycle.yaml",
			wantErr: commerr.ErrInvalidValue,
		},
		{
			name:    "remote ref",
			file:    "remote.yaml",
			wantErr: commerr.ErrInvalidArgument,
		},
		{
			name:    "unknown action",
			file:    "badaction.yaml",
			wantErr: commerr.ErrInvalidValue,
		},
		{name: "missing file", file: "missing.yaml", wantErr: os.ErrNotExist},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := spec.Load(filepath.Join("testdata", tc.file))
			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func TestLoadRejectsMalformedYAML(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "bad.yaml")
	require.NoError(t, os.WriteFile(path, []byte("asyncapi: [unclosed"), 0o600))

	_, err := spec.Load(path)
	require.ErrorIs(t, err, commerr.ErrParseFailed)
}

func TestResolverLookup(t *testing.T) {
	t.Parallel()

	resolver := spec.NewResolver()
	root, err := filepath.Abs(filepath.Join("testdata", "root.yaml"))
	require.NoError(t, err)

	alpha := spec.Location{File: root}.Child("channels").Child("alpha")
	location := alpha.Child("address")
	value, err := resolver.Lookup(location)
	require.NoError(t, err)
	assert.Equal(t, "alpha/{id}", value)
	assert.Equal(t, "address", location.LastToken())

	slashed := spec.Location{File: root}.Child("a/b~c")
	assert.Equal(t, "/a~1b~0c", slashed.Pointer)
	assert.Equal(t, "a/b~c", slashed.LastToken())

	_, err = resolver.Lookup(
		spec.Location{File: root, Pointer: "/servers/nope"},
	)
	require.ErrorIs(t, err, commerr.ErrNotFound)

	_, err = resolver.Lookup(
		spec.Location{File: root, Pointer: "/operations/talk/messages/9"},
	)
	require.ErrorIs(t, err, commerr.ErrNotFound)
}
