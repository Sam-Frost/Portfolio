package spotify

import "testing"

func testKey() []byte {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	return key
}

func TestCipher_EncryptDecryptRoundTrips(t *testing.T) {
	c, err := NewCipher(testKey())
	if err != nil {
		t.Fatalf("NewCipher: %v", err)
	}

	const plaintext = "super-secret-refresh-token"
	ciphertext, err := c.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if ciphertext == plaintext {
		t.Fatal("Encrypt returned the plaintext unchanged")
	}

	got, err := c.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if got != plaintext {
		t.Errorf("Decrypt = %q, want %q", got, plaintext)
	}
}

func TestCipher_EncryptIsNotDeterministic(t *testing.T) {
	c, _ := NewCipher(testKey())

	a, _ := c.Encrypt("token")
	b, _ := c.Encrypt("token")
	if a == b {
		t.Error("two Encrypt calls on the same plaintext produced identical ciphertext (nonce reuse?)")
	}
}

func TestCipher_DecryptRejectsTamperedCiphertext(t *testing.T) {
	c, _ := NewCipher(testKey())

	ciphertext, _ := c.Encrypt("token")
	tampered := ciphertext[:len(ciphertext)-2] + "xx"

	if _, err := c.Decrypt(tampered); err == nil {
		t.Fatal("Decrypt succeeded on tampered ciphertext, want error")
	}
}

func TestCipher_DecryptRejectsWrongKey(t *testing.T) {
	encryptKey := testKey()
	decryptKey := testKey()
	decryptKey[0] ^= 0xFF

	encrypter, _ := NewCipher(encryptKey)
	decrypter, _ := NewCipher(decryptKey)

	ciphertext, _ := encrypter.Encrypt("token")
	if _, err := decrypter.Decrypt(ciphertext); err == nil {
		t.Fatal("Decrypt succeeded with the wrong key, want error")
	}
}
