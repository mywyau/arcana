package chat

import (
	"errors"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	ChannelLocal        = "local"
	MaximumMessageRunes = 200
	MinimumSendInterval = 750 * time.Millisecond
)

var (
	ErrEmptyMessage      = errors.New("enter a message first")
	ErrInvalidChannel    = errors.New("that chat channel is not available")
	ErrInvalidCharacters = errors.New("chat messages cannot contain control characters")
	ErrMessageTooLong    = errors.New("chat messages can contain at most 200 characters")
	ErrRateLimited       = errors.New("you are sending messages too quickly")
)

// Message is the server-authored representation of a chat event. Keeping it
// independent from the wire protocol lets chat move behind another service
// later without coupling moderation and validation to WebSockets.
type Message struct {
	ID         uint64
	Channel    string
	SenderID   string
	SenderName string
	Text       string
	SentAt     time.Time
}

type Service struct {
	mu       sync.Mutex
	nextID   uint64
	lastSent map[string]time.Time
}

func NewService() *Service {
	return &Service{lastSent: make(map[string]time.Time)}
}

func (s *Service) Submit(senderID, senderName, channel, text string) (Message, error) {
	return s.submitAt(senderID, senderName, channel, text, time.Now())
}

func (s *Service) submitAt(senderID, senderName, channel, text string, now time.Time) (Message, error) {
	if channel != ChannelLocal {
		return Message{}, ErrInvalidChannel
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return Message{}, ErrEmptyMessage
	}
	if !utf8.ValidString(text) || utf8.RuneCountInString(text) > MaximumMessageRunes {
		return Message{}, ErrMessageTooLong
	}
	for _, character := range text {
		if unicode.IsControl(character) {
			return Message{}, ErrInvalidCharacters
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if lastSent := s.lastSent[senderID]; !lastSent.IsZero() && now.Sub(lastSent) < MinimumSendInterval {
		return Message{}, ErrRateLimited
	}
	s.nextID++
	s.lastSent[senderID] = now
	return Message{
		ID: s.nextID, Channel: channel, SenderID: senderID,
		SenderName: senderName, Text: text, SentAt: now,
	}, nil
}
