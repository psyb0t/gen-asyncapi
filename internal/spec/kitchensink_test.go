package spec_test

import (
	"path/filepath"
	"testing"

	"github.com/psyb0t/gogen-asyncapi/internal/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func loadKitchensink(t *testing.T) *spec.Document {
	t.Helper()

	document, err := spec.Load(
		filepath.Join(
			"..",
			"..",
			"examples",
			"kitchensink",
			"kitchensink.yaml",
		),
	)
	require.NoError(t, err)

	return document
}

func messageByKey(
	t *testing.T,
	document *spec.Document,
	key string,
) *spec.Message {
	t.Helper()

	for _, message := range document.Messages {
		if message.Key == key {
			return message
		}
	}

	t.Fatalf("message %s not loaded", key)

	return nil
}

func TestKitchensinkServers(t *testing.T) {
	t.Parallel()

	document := loadKitchensink(t)
	require.Len(t, document.Servers, 2)

	production := document.Servers[0]
	assert.Equal(t, "13", production.ProtocolVersion)
	assert.Equal(
		t,
		"wss://{region}.example.com:{port}/v{version}/ws",
		production.URLTemplate(),
	)
	assert.Equal(t, "wss://eu.example.com:443/v2/ws", production.URL())

	require.Len(t, production.Variables, 3)
	assert.Equal(t, "region", production.Variables[0].Name)
	assert.Equal(
		t,
		[]string{"eu", "us"},
		production.Variables[0].Enum,
		"a $ref to components.serverVariables resolves",
	)
	assert.Equal(t, "Deployment region.", production.Variables[0].Description)

	assert.Equal(
		t,
		"mqtt://broker.example.com",
		document.Servers[1].URL(),
		"a $ref to components.servers resolves",
	)
}

func TestKitchensinkMessageTraits(t *testing.T) {
	t.Parallel()

	event := messageByKey(t, loadKitchensink(t), "Event")

	assert.Equal(
		t,
		"event",
		event.Name,
		"a trait never overrides a property the message sets",
	)
	assert.Equal(
		t,
		"application/cloudevents+json",
		event.ContentType,
		"a trait fills properties the message lacks",
	)
	assert.Equal(
		t,
		"/components/schemas/TraceHeaders",
		event.HeadersAt.Pointer,
		"headers from a trait keep their $ref",
	)
	assert.NotNil(t, event.Headers)
}

func TestKitchensinkMessages(t *testing.T) {
	t.Parallel()

	document := loadKitchensink(t)

	command := messageByKey(t, document, "Command")
	assert.Equal(
		t,
		"$message.payload#/commandId",
		command.CorrelationIDLocation,
		"a $ref correlationId resolves",
	)
	assert.Equal(t, "application/json", command.ContentType)

	require.NotNil(t, command.Bindings)
	assert.Equal(
		t,
		[]string{"ws", "kafka"},
		command.Bindings.Keys,
		"a trait adds bindings the message lacks",
	)

	ws := command.Bindings.Object("ws")
	require.NotNil(t, ws)
	assert.Equal(
		t,
		"0.1.0",
		ws.String("bindingVersion"),
		"the message's own nested value wins",
	)
	assert.Equal(
		t,
		"GET",
		ws.String("method"),
		"nested keys only the trait has are merged in",
	)

	result := messageByKey(t, document, "CommandResult")
	assert.Equal(t, "$message.payload#/commandId", result.CorrelationIDLocation)

	heartbeat := messageByKey(t, document, "Heartbeat")
	assert.True(t, heartbeat.Deprecated)
	assert.Equal(
		t,
		"application/json",
		heartbeat.ContentType,
		"defaultContentType applies",
	)

	avro := messageByKey(t, document, "AvroReading")
	assert.Nil(t, avro.Payload)
	assert.Equal(
		t,
		"application/vnd.apache.avro+json;version=1.9.0",
		avro.UnsupportedFormat,
	)

	tree := messageByKey(t, document, "TreeMessage")
	assert.Equal(
		t,
		"/components/schemas/Tree",
		tree.PayloadAt.Pointer,
		"an AsyncAPI multi-format schema unwraps",
	)
	assert.Empty(t, tree.UnsupportedFormat)

	external := messageByKey(t, document, "ExternalPing")
	assert.Equal(t, "external.yaml", filepath.Base(external.Location.File))
}

func TestKitchensinkOperations(t *testing.T) {
	t.Parallel()

	document := loadKitchensink(t)

	operations := map[string]*spec.Operation{}
	for _, operation := range document.Operations {
		operations[operation.Name] = operation
	}

	sendCommand := operations["sendCommand"]
	require.NotNil(t, sendCommand, "a $ref to components.operations resolves")
	assert.Equal(t, "Send a command.", sendCommand.Summary)
	require.NotNil(
		t,
		sendCommand.Reply,
		"a $ref to components.replies resolves",
	)
	assert.Equal(t, "replies", sendCommand.Reply.Channel.Name)
	assert.Equal(
		t,
		"$message.header#/replyTo",
		sendCommand.Reply.AddressLocation,
	)

	onEvent := operations["onEvent"]
	assert.Equal(
		t,
		"Logged operation.",
		onEvent.Summary,
		"an operation trait fills the summary",
	)

	replies := sendCommand.Reply.Channel
	assert.Empty(t, replies.Address, "a null address stays empty")

	componentNames := make([]string, 0, len(document.ComponentSchemas))
	for _, component := range document.ComponentSchemas {
		componentNames = append(componentNames, component.Name)
	}

	assert.Contains(t, componentNames, "Tree")
	assert.Contains(t, componentNames, "Envelope")
}
