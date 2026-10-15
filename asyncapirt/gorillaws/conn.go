// Package gorillaws is an asyncapirt.Transport over gorilla/websocket. Dial
// opens a connection with optional headers, bearer token, query parameters
// and subprotocols, and the returned Conn plugs into a generated Client.
package gorillaws

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	commonhttp "github.com/psyb0t/common-go/http"
	"github.com/psyb0t/ctxerrors"
	"github.com/psyb0t/ctxerrors/commerr"
	"github.com/psyb0t/ctxscope"
)

const closeWriteTimeout = 5 * time.Second

// Conn is one WebSocket connection. One goroutine may read while others
// write; writes are serialized.
type Conn struct {
	conn        *websocket.Conn
	messageType int
	writeMu     sync.Mutex
	closeOnce   sync.Once
	closeErr    error
}

type options struct {
	header       http.Header
	query        url.Values
	subprotocols []string
	messageType  int
	dialer       *websocket.Dialer
	readLimit    int64
}

// Option configures Dial.
type Option func(*options)

// WithHeader adds a request header to the WebSocket handshake.
func WithHeader(name, value string) Option {
	return func(o *options) {
		o.header.Add(name, value)
	}
}

// WithBearerToken sends token in the Authorization header of the handshake.
func WithBearerToken(token string) Option {
	return WithHeader(
		commonhttp.HeaderAuthorization,
		commonhttp.AuthSchemeBearer+token,
	)
}

// WithQueryParam adds a query parameter to the handshake URL, for servers that
// take credentials or options in the query string.
func WithQueryParam(name, value string) Option {
	return func(o *options) {
		o.query.Add(name, value)
	}
}

// WithSubprotocols offers subprotocols in the handshake. Some servers carry a
// bearer token as a subprotocol.
func WithSubprotocols(subprotocols ...string) Option {
	return func(o *options) {
		o.subprotocols = append(o.subprotocols, subprotocols...)
	}
}

// WithBinaryFrames sends messages as binary frames instead of text frames.
func WithBinaryFrames() Option {
	return func(o *options) {
		o.messageType = websocket.BinaryMessage
	}
}

// WithDialer replaces the default dialer, for proxies, TLS settings or
// handshake timeouts.
func WithDialer(dialer *websocket.Dialer) Option {
	return func(o *options) {
		o.dialer = dialer
	}
}

// WithReadLimit caps the size of one incoming message in bytes. Larger
// messages close the connection.
func WithReadLimit(limit int64) Option {
	return func(o *options) {
		o.readLimit = limit
	}
}

// Dial opens a WebSocket connection to rawURL.
func Dial(ctx context.Context, rawURL string, opts ...Option) (*Conn, error) {
	settings := &options{
		header:      http.Header{},
		query:       url.Values{},
		messageType: websocket.TextMessage,
		dialer:      websocket.DefaultDialer,
	}

	for _, opt := range opts {
		opt(settings)
	}

	target, err := handshakeURL(rawURL, settings.query)
	if err != nil {
		return nil, err
	}

	dialer := *settings.dialer
	dialer.Subprotocols = append(
		append([]string{}, dialer.Subprotocols...),
		settings.subprotocols...,
	)

	logger := ctxscope.GetLogger(ctx)

	conn, response, err := dialer.DialContext(
		ctx,
		target.String(),
		settings.header,
	)
	closeHandshakeBody(ctx, response)

	if err != nil {
		return nil, ctxerrors.Wrapf(
			commerr.ErrOpenFailed,
			"dial %s (status %d): %v",
			target.Host,
			handshakeStatus(response),
			err,
		)
	}

	if settings.readLimit > 0 {
		conn.SetReadLimit(settings.readLimit)
	}

	logger.Info(
		"websocket connected",
		"host", target.Host,
		"path", target.Path,
		"subprotocol", conn.Subprotocol(),
	)

	return &Conn{conn: conn, messageType: settings.messageType}, nil
}

// handshakeURL parses rawURL and adds the configured query parameters.
func handshakeURL(rawURL string, extra url.Values) (*url.URL, error) {
	target, err := url.Parse(rawURL)
	if err != nil {
		return nil, ctxerrors.Wrapf(
			commerr.ErrInvalidArgument,
			"websocket url: %v",
			err,
		)
	}

	if len(extra) == 0 {
		return target, nil
	}

	query := target.Query()

	for name, values := range extra {
		for _, value := range values {
			query.Add(name, value)
		}
	}

	target.RawQuery = query.Encode()

	return target, nil
}

// closeHandshakeBody closes the body of a handshake response, when there is
// one.
func closeHandshakeBody(ctx context.Context, response *http.Response) {
	if response == nil || response.Body == nil {
		return
	}

	if err := response.Body.Close(); err != nil {
		ctxscope.GetLogger(ctx).Warn(
			"close handshake response body failed",
			"err", err,
		)
	}
}

// handshakeStatus returns the HTTP status of a failed handshake, or 0 when
// the server never answered.
func handshakeStatus(response *http.Response) int {
	if response == nil {
		return 0
	}

	return response.StatusCode
}

// WriteMessage sends data as one message. The ctx deadline, when set, bounds
// the write.
func (c *Conn) WriteMessage(ctx context.Context, data []byte) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	deadline, _ := ctx.Deadline()
	if err := c.conn.SetWriteDeadline(deadline); err != nil {
		return ctxerrors.Wrap(err, "set write deadline")
	}

	if err := c.conn.WriteMessage(c.messageType, data); err != nil {
		return ctxerrors.Wrapf(
			commerr.ErrWriteFailed,
			"websocket write: %v",
			err,
		)
	}

	return nil
}

// ReadMessage blocks until a message arrives or ctx ends. Text and binary
// messages are both returned. After ctx ends the connection is no longer
// readable, because gorilla/websocket cannot resume a timed-out read.
func (c *Conn) ReadMessage(ctx context.Context) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, ctxerrors.Wrap(err, "read message")
	}

	done := make(chan struct{})
	defer close(done)

	go func() {
		select {
		case <-ctx.Done():
			// An expired deadline is the only way to unblock a gorilla read.
			if err := c.conn.SetReadDeadline(time.Now()); err != nil {
				ctxscope.GetLogger(ctx).Warn(
					"interrupt websocket read failed",
					"err",
					err,
				)
			}
		case <-done:
		}
	}()

	_, data, err := c.conn.ReadMessage()
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, ctxerrors.Wrap(ctxErr, "read message")
		}

		return nil, ctxerrors.Wrapf(
			commerr.ErrReadFailed,
			"websocket read: %v",
			err,
		)
	}

	return data, nil
}

// Close sends a normal close frame and closes the connection. Calling it more
// than once returns the first result.
func (c *Conn) Close() error {
	c.closeOnce.Do(func() {
		c.writeMu.Lock()
		closeFrame := websocket.FormatCloseMessage(
			websocket.CloseNormalClosure,
			"",
		)
		writeErr := c.conn.WriteControl(
			websocket.CloseMessage,
			closeFrame,
			time.Now().Add(closeWriteTimeout),
		)
		c.writeMu.Unlock()

		closeErr := c.conn.Close()

		switch {
		case closeErr != nil:
			c.closeErr = ctxerrors.Wrapf(
				commerr.ErrCloseFailed,
				"websocket close: %v",
				closeErr,
			)
		case writeErr != nil && !errors.Is(writeErr, websocket.ErrCloseSent):
			c.closeErr = ctxerrors.Wrapf(
				commerr.ErrCloseFailed,
				"websocket close frame: %v",
				writeErr,
			)
		}
	})

	return c.closeErr
}
