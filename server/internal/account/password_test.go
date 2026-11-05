package account

import (
	"errors"
	"strings"
	"testing"
)

func TestPasswordHasherHashesAndVerifiesPassword(t *testing.T) {
	hasher := NewPasswordHasher()
	first, err := hasher.Hash("correct horse battery staple")
	if err != nil {
		t.Fatalf("Hash() returned an unexpected error: %v", err)
	}
	second, err := hasher.Hash("correct horse battery staple")
	if err != nil {
		t.Fatalf("second Hash() returned an unexpected error: %v", err)
	}
	if first == second {
		t.Fatal("two password hashes used the same salt")
	}
	if !strings.HasPrefix(first, "$argon2id$v=19$") {
		t.Fatalf("encoded hash = %q", first)
	}
	matched, err := hasher.Verify(first, "correct horse battery staple")
	if err != nil || !matched {
		t.Fatalf("Verify(correct) = %t, error %v", matched, err)
	}
	matched, err = hasher.Verify(first, "wrong password")
	if err != nil || matched {
		t.Fatalf("Verify(wrong) = %t, error %v", matched, err)
	}
}

func TestPasswordHasherRejectsMalformedHash(t *testing.T) {
	matched, err := NewPasswordHasher().Verify("not-a-password-hash", "password")
	if matched || !errors.Is(err, ErrInvalidPasswordHash) {
		t.Fatalf("Verify(malformed) = %t, error %v", matched, err)
	}
}
