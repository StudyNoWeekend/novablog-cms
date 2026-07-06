package crypto

import (
	"testing"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key := "change-this-to-a-random-32-char-secret-key"
	plaintext := "my-super-secret-access-key-12345"

	ciphertext, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	if ciphertext == plaintext {
		t.Fatal("ciphertext should differ from plaintext")
	}

	decrypted, err := Decrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if decrypted != plaintext {
		t.Fatalf("expected %q, got %q", plaintext, decrypted)
	}
}

func TestEncryptProducesDifferentCiphertexts(t *testing.T) {
	key := "change-this-to-a-random-32-char-secret-key"
	plaintext := "same-secret"

	c1, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}
	c2, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}
	if c1 == c2 {
		t.Fatal("two encryptions of the same plaintext should produce different ciphertexts due to random nonce")
	}
}

func TestDecryptWithWrongKeyFails(t *testing.T) {
	key := "change-this-to-a-random-32-char-secret-key"
	wrongKey := "another-different-key-value-xyz"
	plaintext := "my-super-secret-access-key-12345"

	ciphertext, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	_, err = Decrypt(ciphertext, wrongKey)
	if err == nil {
		t.Fatal("Decrypt with wrong key should fail")
	}
}

func TestDecryptInvalidBase64Fails(t *testing.T) {
	_, err := Decrypt("!!!not-base64!!!", "any-key")
	if err == nil {
		t.Fatal("Decrypt of invalid base64 should fail")
	}
}

func TestEncryptDecryptEmptyString(t *testing.T) {
	ciphertext, err := Encrypt("", "any-key")
	if err != nil {
		t.Fatalf("Encrypt of empty string failed: %v", err)
	}
	if ciphertext != "" {
		t.Fatalf("expected empty ciphertext, got %q", ciphertext)
	}

	decrypted, err := Decrypt("", "any-key")
	if err != nil {
		t.Fatalf("Decrypt of empty string failed: %v", err)
	}
	if decrypted != "" {
		t.Fatalf("expected empty plaintext, got %q", decrypted)
	}
}
