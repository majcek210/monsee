package encrypt

import "testing"

var testKey = []byte("01234567890123456789012345678901") // 34 bytes, trimmed below
var key32 = testKey[:32]

func TestEncryptDecryptRoundTrip(t *testing.T) {
	plaintext := "https://discord.com/api/webhooks/secret-path"

	ciphertext, err := Encrypt(key32, plaintext)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if ciphertext == plaintext {
		t.Fatal("ciphertext must not equal plaintext")
	}

	got, err := Decrypt(key32, ciphertext)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if got != plaintext {
		t.Fatalf("got %q, want %q", got, plaintext)
	}
}

func TestEncryptProducesUniqueCiphertexts(t *testing.T) {
	plaintext := "same input every time"

	a, err := Encrypt(key32, plaintext)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	b, err := Encrypt(key32, plaintext)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if a == b {
		t.Fatal("two encryptions of the same plaintext produced identical ciphertext (nonce reuse?)")
	}
}

func TestDecryptWithWrongKeyFails(t *testing.T) {
	ciphertext, err := Encrypt(key32, "top secret")
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	wrongKey := []byte("00000000000000000000000000000000")[:32]
	if _, err := Decrypt(wrongKey, ciphertext); err == nil {
		t.Fatal("expected error decrypting with wrong key, got nil")
	}
}

func TestDecryptInvalidBase64Fails(t *testing.T) {
	if _, err := Decrypt(key32, "not-valid-base64!!!"); err == nil {
		t.Fatal("expected error for invalid base64 input")
	}
}

func TestDecryptTooShortFails(t *testing.T) {
	// "AA==" decodes to a single zero byte — shorter than the GCM nonce.
	if _, err := Decrypt(key32, "AA=="); err == nil {
		t.Fatal("expected error for ciphertext shorter than nonce size")
	}
}

func TestEncryptInvalidKeySizeFails(t *testing.T) {
	shortKey := []byte("too-short")
	if _, err := Encrypt(shortKey, "data"); err == nil {
		t.Fatal("expected error for non-32-byte key")
	}
}
