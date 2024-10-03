package coldmfa

import (
	"slices"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	input := []string{
		"{\"groupName\": \"test\", \"original\": \"otpauth://totp/EphyraSoftware:test-a?algorithm=SHA1&digits=6&issuer=EphyraSoftware&period=30&secret=NL6ZHWZXRNCNNIHQKDXK2Q4GGA3PKQD3\"}",
		"{\"groupName\": \"test\", \"original\": \"otpauth://totp/EphyraSoftware:test-b?algorithm=SHA1&digits=6&issuer=EphyraSoftware&period=30&secret=A23DJ4WDRR2XFPDKBUQ5ZLZN6KVIIIC4\"}",
	}
	slices.Sort(input)

	encrypted, err := EncryptMfaCodeBackupItems(input, "password")
	if err != nil {
		t.Fatal(err)
	}

	decrypted, err := DecryptMfaCodeBackupItems(encrypted, "password")
	if err != nil {
		t.Fatal(err)
	}

	slices.Sort(decrypted)

	if !slices.Equal(input, decrypted) {
		t.Fatalf("expected %v, got %v", input, decrypted)
	}
}

func TestPasswordRequired(t *testing.T) {
	input := []string{
		"{\"groupName\": \"test\", \"original\": \"otpauth://totp/EphyraSoftware:test-a?algorithm=SHA1&digits=6&issuer=EphyraSoftware&period=30&secret=NL6ZHWZXRNCNNIHQKDXK2Q4GGA3PKQD3\"}",
		"{\"groupName\": \"test\", \"original\": \"otpauth://totp/EphyraSoftware:test-b?algorithm=SHA1&digits=6&issuer=EphyraSoftware&period=30&secret=A23DJ4WDRR2XFPDKBUQ5ZLZN6KVIIIC4\"}",
	}
	slices.Sort(input)

	encrypted, err := EncryptMfaCodeBackupItems(input, "password")
	if err != nil {
		t.Fatal(err)
	}

	_, err = DecryptMfaCodeBackupItems(encrypted, "password2")
	if err == nil {
		t.Fatal("expected an error")
	}
}
