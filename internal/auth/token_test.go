package auth

import (
	"fmt"
	"os"
	"runtime"
	"testing"
)

type testPart struct {
	Test string `json:"test"`
}

var wd = getWd()
var wdLen = len(wd) + 1

func getWd() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return wd
}

func location() string {
	if _, file, line, ok := runtime.Caller(2); ok {
		return fmt.Sprintf("%s:%d: ", file[wdLen:], line)
	}
	return ""
}

func equal(t *testing.T, expected any, actual any) {
	if expected != actual {
		t.Errorf("%snot equal\nexpected: %v\nactual: %v", location(), expected, actual)
	}
}
func TestEncodePart(t *testing.T) {
	out, err := encodePart(&testPart{
		Test: "test",
	})
	if err != nil {
		t.Fatal(err)
	}
	equal(t, "eyJ0ZXN0IjoidGVzdCJ9", out)
}

func TestCreateToken(t *testing.T) {
	_, err := createToken()
	if err != nil {
		t.Fatal(err)
	}
}
