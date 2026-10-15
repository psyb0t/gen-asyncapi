package gorillaws_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/psyb0t/ctxerrors/commerr"
	"github.com/psyb0t/gen-asyncapi/asyncapirt/gorillaws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testTimeout = 5 * time.Second

type handshake struct {
	authorization string
	token         string
	subprotocol   string
}

// newEchoServer starts a WebSocket server that reports each handshake and
// echoes every message back with the same frame type.
func newEchoServer(t *testing.T) (string, <-chan handshake) {
	t.Helper()

	handshakes := make(chan handshake, 1)
	upgrader := websocket.Upgrader{Subprotocols: []string{"chat.v1"}}

	server := httptest.NewServer(http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		handshakes <- handshake{
			authorization: r.Header.Get("Authorization"),
			token:         r.URL.Query().Get("token"),
			subprotocol:   strings.Join(websocket.Subprotocols(r), ","),
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		defer func() {
			assert.NoError(t, conn.Close())
		}()

		for {
			messageType, data, err := conn.ReadMessage()
			if err != nil {
				return
			}

			if err := conn.WriteMessage(messageType, data); err != nil {
				return
			}
		}
	}))
	t.Cleanup(server.Close)

	return "ws" + strings.TrimPrefix(server.URL, "http"), handshakes
}

func TestDialSendsCredentialsAndEchoes(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	t.Cleanup(cancel)

	url, handshakes := newEchoServer(t)

	conn, err := gorillaws.Dial(
		ctx,
		url+"/ws?existing=1",
		gorillaws.WithBearerToken("secret"),
		gorillaws.WithQueryParam("token", "q"),
		gorillaws.WithSubprotocols("chat.v1"),
		gorillaws.WithReadLimit(1024),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, conn.Close())
	})

	seen := <-handshakes
	assert.Equal(t, "Bearer secret", seen.authorization)
	assert.Equal(t, "q", seen.token)
	assert.Equal(t, "chat.v1", seen.subprotocol)

	require.NoError(t, conn.WriteMessage(ctx, []byte(`{"a":1}`)))

	data, err := conn.ReadMessage(ctx)
	require.NoError(t, err)
	assert.JSONEq(t, `{"a":1}`, string(data))
}

func TestBinaryFrames(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	t.Cleanup(cancel)

	url, handshakes := newEchoServer(t)

	conn, err := gorillaws.Dial(ctx, url, gorillaws.WithBinaryFrames())
	require.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, conn.Close())
	})

	<-handshakes

	require.NoError(t, conn.WriteMessage(ctx, []byte{0, 1, 2}))

	data, err := conn.ReadMessage(ctx)
	require.NoError(t, err)
	assert.Equal(t, []byte{0, 1, 2}, data)
}

func TestReadMessageStopsWhenContextEnds(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	t.Cleanup(cancel)

	url, handshakes := newEchoServer(t)

	conn, err := gorillaws.Dial(ctx, url)
	require.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, conn.Close())
	})

	<-handshakes

	readCtx, stopRead := context.WithCancel(ctx)
	readErr := make(chan error, 1)

	go func() {
		_, err := conn.ReadMessage(readCtx)
		readErr <- err
	}()

	stopRead()
	require.ErrorIs(t, <-readErr, context.Canceled)

	_, err = conn.ReadMessage(readCtx)
	require.ErrorIs(t, err, context.Canceled)
}

func TestCloseIsIdempotent(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	t.Cleanup(cancel)

	url, handshakes := newEchoServer(t)

	conn, err := gorillaws.Dial(ctx, url)
	require.NoError(t, err)

	<-handshakes

	require.NoError(t, conn.Close())
	require.NoError(t, conn.Close())

	require.ErrorIs(
		t,
		conn.WriteMessage(ctx, []byte("x")),
		commerr.ErrWriteFailed,
	)
}

func TestDialErrors(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	t.Cleanup(cancel)

	_, err := gorillaws.Dial(ctx, "ws://bad host")
	require.ErrorIs(t, err, commerr.ErrInvalidArgument)

	plain := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(plain.Close)

	_, err = gorillaws.Dial(
		ctx,
		"ws"+strings.TrimPrefix(plain.URL, "http"),
		gorillaws.WithHeader("X-Test", "1"),
		gorillaws.WithDialer(&websocket.Dialer{HandshakeTimeout: time.Second}),
	)
	require.ErrorIs(t, err, commerr.ErrOpenFailed)
}
