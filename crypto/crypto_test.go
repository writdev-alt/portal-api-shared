package crypto

import (
	"testing"
)

func TestHashAndSalt(t *testing.T) {
	plainPassword := []byte("test-password-123")

	hash := HashAndSalt(plainPassword)

	if hash == "" {
		t.Error("HashAndSalt returned empty string")
	}

	if hash == string(plainPassword) {
		t.Error("HashAndSalt returned the same string as input")
	}

	if len(hash) < 10 {
		t.Error("HashAndSalt returned hash that is too short")
	}
}

func TestComparePassword(t *testing.T) {
	plainPassword := []byte("test-password-123")

	// Hash the password
	hashedPassword := HashAndSalt(plainPassword)

	// Test correct password
	if !ComparePassword(hashedPassword, plainPassword) {
		t.Error("ComparePassword failed for correct password")
	}

	// Test incorrect password
	wrongPassword := []byte("wrong-password")
	if ComparePassword(hashedPassword, wrongPassword) {
		t.Error("ComparePassword returned true for incorrect password")
	}

	// Test with different hash
	differentHash := HashAndSalt([]byte("different-password"))
	if ComparePassword(differentHash, plainPassword) {
		t.Error("ComparePassword returned true when comparing different hash")
	}
}

func TestHashAndSalt_UniqueHashes(t *testing.T) {
	plainPassword := []byte("same-password")

	hash1 := HashAndSalt(plainPassword)
	hash2 := HashAndSalt(plainPassword)

	// Each hash should be unique due to salt
	if hash1 == hash2 {
		t.Error("HashAndSalt returned same hash for same password (should be salted)")
	}

	// But both should validate against the same password
	if !ComparePassword(hash1, plainPassword) {
		t.Error("First hash does not validate against original password")
	}

	if !ComparePassword(hash2, plainPassword) {
		t.Error("Second hash does not validate against original password")
	}
}

func TestComparePassword_EmptyPassword(t *testing.T) {
	emptyPassword := []byte("")
	hash := HashAndSalt(emptyPassword)

	if !ComparePassword(hash, emptyPassword) {
		t.Error("ComparePassword failed for empty password")
	}

	if ComparePassword(hash, []byte("not-empty")) {
		t.Error("ComparePassword returned true for non-empty password with empty hash")
	}
}

func BenchmarkHashAndSalt(b *testing.B) {
	plainPassword := []byte("benchmark-password-123")

	for i := 0; i < b.N; i++ {
		HashAndSalt(plainPassword)
	}
}

func BenchmarkComparePassword(b *testing.B) {
	plainPassword := []byte("benchmark-password-123")
	hashedPassword := HashAndSalt(plainPassword)

	for i := 0; i < b.N; i++ {
		ComparePassword(hashedPassword, plainPassword)
	}
}
