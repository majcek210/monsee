package hash

import "testing"

func TestPasswordRoundTrip(t *testing.T) {
	hashed, err := Password("correct-horse-battery-staple")
	if err != nil {
		t.Fatalf("Password: %v", err)
	}
	if hashed == "correct-horse-battery-staple" {
		t.Fatal("hash must not equal the plaintext password")
	}

	if !CheckPassword(hashed, "correct-horse-battery-staple") {
		t.Fatal("CheckPassword should succeed with the correct password")
	}
}

func TestCheckPasswordWrongPassword(t *testing.T) {
	hashed, err := Password("correct-horse-battery-staple")
	if err != nil {
		t.Fatalf("Password: %v", err)
	}

	if CheckPassword(hashed, "wrong-password") {
		t.Fatal("CheckPassword should fail with an incorrect password")
	}
}

func TestCheckPasswordMalformedHash(t *testing.T) {
	if CheckPassword("not-a-bcrypt-hash", "anything") {
		t.Fatal("CheckPassword should fail for a malformed hash")
	}
}

func TestSHA256KnownValue(t *testing.T) {
	// sha256("") = e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
	const want = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	if got := SHA256(""); got != want {
		t.Fatalf("SHA256(\"\") = %q, want %q", got, want)
	}
}

func TestSHA256Deterministic(t *testing.T) {
	a := SHA256("sk_abcdef1234567890")
	b := SHA256("sk_abcdef1234567890")
	if a != b {
		t.Fatal("SHA256 should be deterministic for the same input")
	}

	c := SHA256("sk_different_key")
	if a == c {
		t.Fatal("SHA256 of different inputs should differ")
	}
}
