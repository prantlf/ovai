package auth

import (
	"testing"

	"github.com/prantlf/ovai/internal/test"
)

type testPart struct {
	Test string `json:"test"`
}

func TestEncodePart(t *testing.T) {
	out, err := encodePart(&testPart{
		Test: "test",
	})
	if err != nil {
		t.Fatal(err)
	}
	test.Equal(t, "eyJ0ZXN0IjoidGVzdCJ9", out)
}

func TestCreateToken(t *testing.T) {
	_, err := createToken()
	if err != nil {
		t.Fatal(err)
	}
}
