package realtime

import (
	"testing"

	"github.com/zishang520/socket.io/v2/socket"
)

func TestReplyDeliversTheResponseWhenTheClientAskedForOne(t *testing.T) {
	var received []any

	ack := socket.Ack(func(args []any, err error) {
		if err != nil {
			t.Fatalf("ack received error %v", err)
		}
		received = args
	})

	replyTo(ack)(errorAck{Ok: false, Message: "nope"})

	if len(received) != 1 {
		t.Fatalf("ack received %d arguments", len(received))
	}
	if _, ok := received[0].(errorAck); !ok {
		t.Fatalf("ack received %#v", received[0])
	}
}

func TestReplyStaysSilentInsteadOfPanickingWhenTheAckIsAbsentOrForged(t *testing.T) {
	forged := []any{
		nil,
		"not-a-function",
		42,
		map[string]any{},
		[]any{},
		true,
		float64(0),
	}

	for _, last := range forged {
		args := []any{map[string]any{"roomCode": "ABC123"}, last}

		payload, ack := splitArgs(args)
		if ack != nil {
			t.Fatalf("forged ack %#v was accepted", last)
		}
		if payload == nil {
			t.Fatal("payload was dropped along with the forged ack")
		}

		reply := replyTo(ack)
		reply(errorAck{Ok: false, Message: "nope"})
	}
}

func TestReplyIgnoresAnAckWhenTheClientSentNoPayload(t *testing.T) {
	delivered := false

	ack := socket.Ack(func([]any, error) { delivered = true })

	payload, parsed := splitArgs([]any{ack})
	if payload != nil {
		t.Fatalf("payload = %#v", payload)
	}
	if parsed == nil {
		t.Fatal("ack was not recognized")
	}

	replyTo(parsed)(plainAck{Ok: true})

	if !delivered {
		t.Fatal("ack was never invoked")
	}
}

func TestSplitArgsHandlesAnEmptyArgumentList(t *testing.T) {
	payload, ack := splitArgs(nil)
	if payload != nil || ack != nil {
		t.Fatalf("splitArgs(nil) = %#v, %v", payload, ack)
	}
}

func TestFieldReadsOnlyObjectPayloads(t *testing.T) {
	payload := map[string]any{"roomCode": "ABC123"}

	if got := field(payload, "roomCode"); got != "ABC123" {
		t.Fatalf("field = %v", got)
	}
	if got := field(payload, "missing"); got != nil {
		t.Fatalf("field = %v", got)
	}

	for _, broken := range []any{nil, "ABC123", 42, []any{"ABC123"}, true} {
		if got := field(broken, "roomCode"); got != nil {
			t.Fatalf("field(%#v) = %v", broken, got)
		}
	}
}
