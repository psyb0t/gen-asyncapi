package chat_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/psyb0t/ctxerrors"
	"github.com/psyb0t/ctxerrors/commerr"
	"github.com/psyb0t/gen-asyncapi/asyncapirt"
	"github.com/psyb0t/gen-asyncapi/examples/chat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testTimeout = 5 * time.Second

func TestJoinRequestMarshalFillsConst(t *testing.T) {
	t.Parallel()

	data, err := json.Marshal(chat.JoinRequest{ID: "1", Room: "lobby"})
	require.NoError(t, err)

	assert.JSONEq(t, `{"type":"join","id":"1","room":"lobby"}`, string(data))
}

func TestJoinRequestUnmarshal(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		input   string
		wantErr error
	}{
		{
			name:  "valid with optional nickname",
			input: `{"type":"join","id":"1","room":"lobby","nickname":"neo"}`,
		},
		{
			name:    "missing required room",
			input:   `{"type":"join","id":"1"}`,
			wantErr: commerr.ErrRequiredFieldNotSet,
		},
		{
			name:    "undeclared property rejected",
			input:   `{"type":"join","id":"1","room":"lobby","extra":true}`,
			wantErr: commerr.ErrValidationFailed,
		},
		{
			name:    "wrong const",
			input:   `{"type":"leave","id":"1","room":"lobby"}`,
			wantErr: commerr.ErrValidationFailed,
		},
		{
			name:    "pattern mismatch",
			input:   `{"type":"join","id":"1","room":"Lobby!"}`,
			wantErr: commerr.ErrValidationFailed,
		},
		{
			name:    "empty room below minLength",
			input:   `{"type":"join","id":"1","room":""}`,
			wantErr: commerr.ErrValidationFailed,
		},
		{
			name:    "not an object",
			input:   `["join"]`,
			wantErr: commerr.ErrCouldNotDecode,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var request chat.JoinRequest

			err := json.Unmarshal([]byte(tc.input), &request)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, "lobby", request.Room)
			require.NotNil(t, request.Nickname)
			assert.Equal(t, "neo", *request.Nickname)
		})
	}
}

func TestSayRequestValidate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		request  chat.SayRequest
		wantFail bool
	}{
		{
			name: "valid",
			request: chat.SayRequest{
				ID:       "1",
				Text:     "hi",
				Mentions: []string{"a"},
			},
		},
		{name: "empty text", request: chat.SayRequest{ID: "1"}, wantFail: true},
		{
			name: "too many mentions",
			request: chat.SayRequest{
				ID:       "1",
				Text:     "hi",
				Mentions: []string{"a", "b", "c", "d", "e", "f"},
			},
			wantFail: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.request.Validate()
			if tc.wantFail {
				require.ErrorIs(t, err, commerr.ErrValidationFailed)

				return
			}

			require.NoError(t, err)
		})
	}
}

func TestServerEventDecode(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		input     string
		wantValue any
		wantErr   error
	}{
		{
			name: "explicit mapping selects MessagePosted",
			input: `{"type":"message.posted","at":"2026-01-02T03:04:05Z","sequence":7,` +
				`"text":"hi","author":{"id":"u1","role":"owner"}}`,
			wantValue: chat.MessagePosted{
				At: time.Date(
					2026,
					time.January,
					2,
					3,
					4,
					5,
					0,
					time.UTC,
				),
				Sequence: 7,
				Type:     "message.posted",
				Text:     "hi",
				Author:   chat.Member{ID: "u1", Role: chat.MemberRoleOwner},
			},
		},
		{
			name: "const selects MemberJoined",
			input: `{"type":"member.joined","at":"2026-01-02T03:04:05Z","sequence":8,` +
				`"member":{"id":"u2","role":"guest"}}`,
			wantValue: chat.MemberJoined{
				At: time.Date(
					2026,
					time.January,
					2,
					3,
					4,
					5,
					0,
					time.UTC,
				),
				Sequence: 8,
				Type:     "member.joined",
				Member:   chat.Member{ID: "u2", Role: chat.MemberRoleGuest},
			},
		},
		{
			name:      "const selects Ping",
			input:     `{"type":"ping"}`,
			wantValue: chat.Ping{Type: "ping"},
		},
		{
			name:    "unknown discriminator",
			input:   `{"type":"bogus"}`,
			wantErr: asyncapirt.ErrUnknownDiscriminator,
		},
		{
			name:    "missing discriminator",
			input:   `{"text":"hi"}`,
			wantErr: commerr.ErrRequiredFieldNotSet,
		},
		{
			name: "invalid nested enum",
			input: `{"type":"member.joined","at":"2026-01-02T03:04:05Z","sequence":8,` +
				`"member":{"id":"u2","role":"admin"}}`,
			wantErr: commerr.ErrValidationFailed,
		},
		{
			name:    "sequence below minimum",
			input:   `{"type":"member.joined","at":"2026-01-02T03:04:05Z","sequence":0,"member":{"id":"u2","role":"guest"}}`,
			wantErr: commerr.ErrValidationFailed,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var event chat.ServerEvent

			err := json.Unmarshal([]byte(tc.input), &event)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)

			value, err := event.Value()
			require.NoError(t, err)
			assert.Equal(t, tc.wantValue, value)
		})
	}
}

func TestServerEventFromAndAs(t *testing.T) {
	t.Parallel()

	var event chat.ServerEvent
	require.NoError(t, event.FromPing(chat.Ping{}))

	data, err := json.Marshal(event)
	require.NoError(t, err)
	assert.JSONEq(t, `{"type":"ping"}`, string(data))

	discriminator, err := event.Discriminator()
	require.NoError(t, err)
	assert.Equal(t, "ping", discriminator)

	_, err = event.AsMemberJoined()
	require.Error(t, err)

	err = event.FromMemberJoined(
		chat.MemberJoined{
			Sequence: 0,
			Member:   chat.Member{ID: "u", Role: chat.MemberRoleGuest},
		},
	)
	require.ErrorIs(t, err, commerr.ErrValidationFailed)

	var empty chat.ServerEvent

	data, err = json.Marshal(empty)
	require.NoError(t, err)
	assert.Equal(t, "null", string(data))
}

func TestAttachmentMetaValue(t *testing.T) {
	t.Parallel()

	attachment := chat.Attachment{Kind: chat.AttachmentKindImage}
	mime := "image/png"
	require.NoError(
		t,
		attachment.SetMetaValue(chat.AttachmentMeta{Size: 42, Mime: &mime}),
	)
	assert.JSONEq(t, `{"size":42,"mime":"image/png"}`, attachment.Meta)

	meta, err := attachment.MetaValue()
	require.NoError(t, err)
	assert.Equal(t, int64(42), meta.Size)

	attachment.Meta = `{"size":-1}`
	_, err = attachment.MetaValue()
	require.ErrorIs(t, err, commerr.ErrValidationFailed)

	attachment.Meta = `not json`
	_, err = attachment.MetaValue()
	require.ErrorIs(t, err, commerr.ErrCouldNotDecode)
}

func TestDecodeReceivedMessage(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    string
		wantName string
		wantErr  error
	}{
		{
			name:     "ack pinned by const",
			input:    `{"type":"ack","id":"1","ok":true}`,
			wantName: chat.MessageAck,
		},
		{
			name:     "event by structure",
			input:    `{"type":"ping"}`,
			wantName: chat.MessageEvent,
		},
		{
			name:     "notice by config rule",
			input:    `{"messages":[{"type":"notice","text":"maintenance"}]}`,
			wantName: chat.MessageNotice,
		},
		{
			name:     "notice by the second value of a list rule",
			input:    `{"messages":[{"type":"info","text":"maintenance"}]}`,
			wantName: chat.MessageNotice,
		},
		{
			name:     "alert by config rule",
			input:    `{"messages":[{"type":"alert","text":"overload"}]}`,
			wantName: chat.MessageAlert,
		},
		{
			name:    "unknown shape",
			input:   `{"hello":"world"}`,
			wantErr: asyncapirt.ErrUnknownMessage,
		},
		{
			name:    "pinned but invalid",
			input:   `{"type":"ack","id":"1"}`,
			wantErr: asyncapirt.ErrUnknownMessage,
		},
		{name: "not json", input: `{`, wantErr: commerr.ErrCouldNotDecode},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			message, err := chat.DecodeReceivedMessage([]byte(tc.input))
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.wantName, message.Name)
		})
	}
}

func TestRoomAddress(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "rooms/lobby", chat.RoomAddress("lobby"))
	assert.Equal(t, "rooms/{roomId}", chat.RoomAddressTemplate)
	assert.Equal(t, "wss://chat.example.com/ws", chat.ServerProductionURL)
}

// memoryTransport is an in-process asyncapirt.Transport.
type memoryTransport struct {
	incoming chan []byte
	written  chan []byte
	closed   chan struct{}
}

func newMemoryTransport() *memoryTransport {
	return &memoryTransport{
		incoming: make(chan []byte, 8),
		written:  make(chan []byte, 8),
		closed:   make(chan struct{}),
	}
}

func (m *memoryTransport) WriteMessage(_ context.Context, data []byte) error {
	m.written <- data

	return nil
}

func (m *memoryTransport) ReadMessage(ctx context.Context) ([]byte, error) {
	select {
	case data := <-m.incoming:
		return data, nil
	case <-ctx.Done():
		return nil, ctxerrors.Wrap(ctx.Err(), "read message")
	}
}

func (m *memoryTransport) Close() error {
	close(m.closed)

	return nil
}

// recordingHandler sends every message it gets to a channel.
type recordingHandler struct {
	chat.UnimplementedHandler

	acks     chan chat.Ack
	events   chan chat.ServerEvent
	unknowns chan error
}

func (h recordingHandler) HandleAck(_ context.Context, payload chat.Ack) error {
	h.acks <- payload

	return nil
}

func (h recordingHandler) HandleEvent(
	_ context.Context,
	payload chat.ServerEvent,
) error {
	h.events <- payload

	return nil
}

func (h recordingHandler) HandleUnknownMessage(
	_ context.Context,
	_ []byte,
	err error,
) error {
	h.unknowns <- err

	return nil
}

func TestClientSend(t *testing.T) {
	t.Parallel()

	transport := newMemoryTransport()
	client := chat.NewClient(transport)

	require.NoError(
		t,
		client.JoinRoom(
			context.Background(),
			chat.JoinRequest{ID: "1", Room: "lobby"},
		),
	)
	assert.JSONEq(
		t,
		`{"type":"join","id":"1","room":"lobby"}`,
		string(<-transport.written),
	)

	err := client.JoinRoom(
		context.Background(),
		chat.JoinRequest{ID: "1", Room: "NOT VALID"},
	)
	require.ErrorIs(t, err, commerr.ErrValidationFailed)
	assert.Empty(t, transport.written)

	require.NoError(
		t,
		client.Say(
			context.Background(),
			chat.SayRequest{ID: "2", Text: "hello"},
		),
	)
	assert.JSONEq(
		t,
		`{"type":"say","id":"2","text":"hello"}`,
		string(<-transport.written),
	)

	require.NoError(t, client.Close())

	select {
	case <-transport.closed:
	default:
		t.Fatal("transport was not closed")
	}
}

func TestClientRun(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	t.Cleanup(cancel)

	transport := newMemoryTransport()
	handler := recordingHandler{
		acks:     make(chan chat.Ack, 1),
		events:   make(chan chat.ServerEvent, 1),
		unknowns: make(chan error, 1),
	}

	runErr := make(chan error, 1)

	go func() {
		runErr <- chat.NewClient(transport).Run(ctx, handler)
	}()

	transport.incoming <- []byte(`{"type":"ack","id":"1","ok":true}`)

	ack := <-handler.acks
	assert.True(t, ack.Ok)

	transport.incoming <- []byte(`{"type":"ping"}`)

	event := <-handler.events
	discriminator, err := event.Discriminator()
	require.NoError(t, err)
	assert.Equal(t, "ping", discriminator)

	transport.incoming <- []byte(`{"nope":1}`)

	require.ErrorIs(t, <-handler.unknowns, asyncapirt.ErrUnknownMessage)

	transport.incoming <- []byte(`{"messages":[{"type":"notice","text":"x"}]}`)

	cancel()
	require.ErrorIs(t, <-runErr, context.Canceled)
}

func TestDispatchRejectsUnknownName(t *testing.T) {
	t.Parallel()

	err := chat.Dispatch(
		context.Background(),
		chat.UnimplementedHandler{},
		chat.ReceivedMessage{Name: "bogus"},
	)
	require.ErrorIs(t, err, asyncapirt.ErrUnknownMessage)

	err = chat.Dispatch(
		context.Background(),
		chat.UnimplementedHandler{},
		chat.ReceivedMessage{Name: chat.MessageAck, Payload: 1},
	)
	require.ErrorIs(t, err, commerr.ErrInvalidValue)
}
