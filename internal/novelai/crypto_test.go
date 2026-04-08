package novelai

import "testing"

func TestDeriveKeysIsDeterministic(t *testing.T) {
	a1, e1, err := DeriveKeys("user@example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}
	a2, e2, err := DeriveKeys("user@example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}
	if a1 != a2 || e1 != e2 {
		t.Fatalf("keys are not deterministic")
	}
}

func TestDeriveKeysNormalizesEmail(t *testing.T) {
	a1, e1, err := DeriveKeys(" User@Example.com ", "password123")
	if err != nil {
		t.Fatal(err)
	}
	a2, e2, err := DeriveKeys("user@example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}
	if a1 != a2 || e1 != e2 {
		t.Fatalf("email normalization mismatch")
	}
}

func TestDeriveKeysShape(t *testing.T) {
	accessKey, encryptionKey, accessSalt, encryptionSalt, err := DeriveKeyDebug("user@example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}
	if len(accessKey) != 64 {
		t.Fatalf("access key length = %d", len(accessKey))
	}
	if len(encryptionKey) == 0 {
		t.Fatal("expected encryption key")
	}
	if len(accessSalt) != 32 || len(encryptionSalt) != 32 {
		t.Fatalf("unexpected salt lengths: access=%d encryption=%d", len(accessSalt), len(encryptionSalt))
	}
	if accessSalt == encryptionSalt {
		t.Fatal("expected distinct salts")
	}
}
