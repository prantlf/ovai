package auth

import (
	"testing"
)

func TestCreateToken(t *testing.T) {
	_, err := createToken()
	if err != nil {
		t.Fatal(err)
	}
}
