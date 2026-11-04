package protocol

import "testing"

func TestDecodeClientMessage(t *testing.T) {
	message, err := DecodeClientMessage([]byte(`{"type":"move_to","sequence":3,"destination":{"x":1.5,"y":4,"z":-2}}`))
	if err != nil {
		t.Fatalf("DecodeClientMessage() returned an unexpected error: %v", err)
	}
	if message.Type != MessageMoveTo || message.Sequence != 3 {
		t.Fatalf("DecodeClientMessage() = %+v", message)
	}
	if message.Destination == nil || message.Destination.X != 1.5 || message.Destination.Y != 4 || message.Destination.Z != -2 {
		t.Fatalf("DecodeClientMessage().Destination = %+v", message.Destination)
	}
}

func TestDecodeEngageTargetMessage(t *testing.T) {
	message, err := DecodeClientMessage([]byte(`{"type":"engage_target","sequence":2,"target_id":"dummy"}`))
	if err != nil {
		t.Fatalf("DecodeClientMessage() returned an unexpected error: %v", err)
	}
	if message.Type != MessageEngageTarget || message.Sequence != 2 || message.TargetID != "dummy" {
		t.Fatalf("DecodeClientMessage() = %+v", message)
	}
}

func TestDecodeDisengageMessage(t *testing.T) {
	message, err := DecodeClientMessage([]byte(`{"type":"disengage","sequence":3}`))
	if err != nil {
		t.Fatalf("DecodeClientMessage() returned an unexpected error: %v", err)
	}
	if message.Type != MessageDisengage || message.Sequence != 3 {
		t.Fatalf("DecodeClientMessage() = %+v", message)
	}
}

func TestDecodeClientMessageRejectsInvalidMessages(t *testing.T) {
	tests := map[string]string{
		"malformed JSON":    `{"type":`,
		"unknown field":     `{"type":"move_to","sequence":1,"destination":{"x":1,"z":2},"extra":true}`,
		"unknown type":      `{"type":"dance","sequence":1,"destination":{"x":1,"z":2}}`,
		"missing sequence":  `{"type":"move_to","destination":{"x":1,"z":2}}`,
		"missing target":    `{"type":"move_to","sequence":1}`,
		"missing target ID": `{"type":"engage_target","sequence":1}`,
		"multiple values":   `{"type":"move_to","sequence":1,"destination":{"x":1,"z":2}} {}`,
	}

	for name, input := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := DecodeClientMessage([]byte(input)); err == nil {
				t.Fatal("DecodeClientMessage() error = nil")
			}
		})
	}
}
