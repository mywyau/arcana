package chat

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestSubmitNormalizesAndNumbersMessages(t *testing.T) {
	service := NewService()
	now := time.Date(2026, time.November, 18, 12, 0, 0, 0, time.UTC)
	first, err := service.submitAt("one", "Player One", ChannelLocal, "  hello world  ", now)
	if err != nil {
		t.Fatalf("submitAt() error = %v", err)
	}
	second, err := service.submitAt("two", "Player Two", ChannelLocal, "hello back", now)
	if err != nil {
		t.Fatalf("submitAt() second error = %v", err)
	}
	if first.ID != 1 || second.ID != 2 || first.Text != "hello world" || first.SenderName != "Player One" {
		t.Fatalf("messages = %+v, %+v", first, second)
	}
}

func TestSubmitRejectsInvalidMessages(t *testing.T) {
	tests := []struct {
		name    string
		channel string
		text    string
		want    error
	}{
		{name: "channel", channel: "world", text: "hello", want: ErrInvalidChannel},
		{name: "empty", channel: ChannelLocal, text: "   ", want: ErrEmptyMessage},
		{name: "control", channel: ChannelLocal, text: "hello\nworld", want: ErrInvalidCharacters},
		{name: "long", channel: ChannelLocal, text: strings.Repeat("界", MaximumMessageRunes+1), want: ErrMessageTooLong},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewService().submitAt("player", "Player", test.channel, test.text, time.Now())
			if !errors.Is(err, test.want) {
				t.Fatalf("submitAt() error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestSubmitRateLimitsPerSender(t *testing.T) {
	service := NewService()
	now := time.Now()
	if _, err := service.submitAt("one", "One", ChannelLocal, "first", now); err != nil {
		t.Fatal(err)
	}
	if _, err := service.submitAt("one", "One", ChannelLocal, "second", now.Add(MinimumSendInterval-time.Millisecond)); !errors.Is(err, ErrRateLimited) {
		t.Fatalf("early submit error = %v, want %v", err, ErrRateLimited)
	}
	if _, err := service.submitAt("two", "Two", ChannelLocal, "independent", now); err != nil {
		t.Fatalf("other sender error = %v", err)
	}
	if _, err := service.submitAt("one", "One", ChannelLocal, "ready", now.Add(MinimumSendInterval)); err != nil {
		t.Fatalf("ready submit error = %v", err)
	}
}
