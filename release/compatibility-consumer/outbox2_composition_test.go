package compatibilityconsumer_test

import (
	"errors"
	"net/url"
	"testing"
	"time"

	eventoutbox "github.com/faustbrian/go-event-sourcing/adapters/outbox/v3"
	eventsourcing "github.com/faustbrian/go-event-sourcing/v2"
	outbox "github.com/faustbrian/go-transactional-outbox/v2"
	"github.com/faustbrian/go-transactional-outbox/v2/relay"
	webhook "github.com/faustbrian/go-webhook/v3"
	webhookoutbox "github.com/faustbrian/go-webhook/v3/adapters/outbox"
)

var (
	_ relay.Publisher                                      = (*webhookoutbox.Publisher)(nil)
	_ func(outbox.Envelope) (eventsourcing.Message, error) = (*eventoutbox.EnvelopeCodec)(nil).Decode
)

func TestPublicOutbox2Webhook3Composition(t *testing.T) {
	now := time.Unix(1_700_000_000, 0).UTC()
	builder, err := outbox.NewEnvelopeBuilder(
		outbox.WithClock(func() time.Time { return now }),
		outbox.WithIDGenerator(func() (string, error) { return "envelope-1", nil }),
	)
	if err != nil {
		t.Fatal(err)
	}
	endpoint, err := url.Parse("https://example.com/hook")
	if err != nil {
		t.Fatal(err)
	}
	delivery := webhook.DeliveryRequest{
		Endpoint: endpoint, Body: []byte("body"), EventID: "event-1",
		IdempotencyKey: "command-1", Metadata: map[string]string{"tenant": "one"},
	}
	envelope, err := webhookoutbox.Build(builder, "webhooks", delivery, 4096)
	if err != nil {
		t.Fatal(err)
	}
	if envelope.ID != "envelope-1" || envelope.Topic != "webhooks" ||
		envelope.OrderingKey != "event-1" || envelope.IdempotencyKey != "command-1" {
		t.Fatal("published envelope routing identity changed")
	}
	delivery.Body[0] = 'x'
	delivery.Metadata["tenant"] = "changed"
	decoded, err := webhook.UnmarshalDeliveryRequest(envelope.Payload, 4096)
	if err != nil || string(decoded.Body) != "body" || envelope.Metadata["tenant"] != "one" {
		t.Fatal("published delivery input ownership changed")
	}
	if _, err := webhookoutbox.Build(builder, "webhooks", delivery, 1); !errors.Is(err, webhook.ErrDeliveryEncoding) {
		t.Fatalf("delivery encoding budget bypassed: %v", err)
	}
}

func TestPublicOutbox2EventOutbox3Composition(t *testing.T) {
	stream, err := eventsourcing.NewStreamID("account", "account-1")
	if err != nil {
		t.Fatal(err)
	}
	event, err := eventsourcing.NewEncodedEvent(eventsourcing.EncodedEventInput{
		Name: "account.opened", Version: 1, ContentType: "application/json",
		Payload: []byte(`{"owner":"Ada"}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	pending, err := eventsourcing.NewPendingMessage(eventsourcing.PendingMessageInput{
		ID: "message-1", Stream: stream, Event: event,
		RecordedAt: time.Unix(1_700_000_000, 0).UTC(),
	})
	if err != nil {
		t.Fatal(err)
	}
	message, err := eventsourcing.NewMessage(eventsourcing.MessageInput{
		Pending: pending, StreamVersion: 1, GlobalPosition: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	codec, err := eventoutbox.NewEnvelopeCodec(eventoutbox.FixedTopic("events"), outbox.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	envelope, err := codec.Encode(message)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := codec.Decode(envelope)
	if err != nil || !decoded.Equal(message) {
		t.Fatal("published event envelope round trip changed")
	}
	envelope.Metadata[eventoutbox.MetadataMessageID] = "different-message"
	if _, err := codec.Decode(envelope); !errors.Is(err, eventoutbox.ErrEnvelopeCorrupt) {
		t.Fatalf("inconsistent envelope identity accepted: %v", err)
	}
}
