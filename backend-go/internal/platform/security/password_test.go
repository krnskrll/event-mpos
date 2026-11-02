package security

import "testing"

func TestPasswordHashAndCheck(t *testing.T) {
	hash, err := HashPassword("Admin123!")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if hash == "Admin123!" {
		t.Fatal("password stored in plaintext")
	}
	if !CheckPassword(hash, "Admin123!") {
		t.Error("correct password rejected")
	}
	if CheckPassword(hash, "wrong") {
		t.Error("wrong password accepted")
	}
}
