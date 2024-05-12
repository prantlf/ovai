package test

import (
	"fmt"
	"os"
	"runtime"
	"testing"
)

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

func Equal(t *testing.T, expected any, actual any) {
	if expected != actual {
		t.Errorf("%snot equal\nexpected: %v\nactual: %v", location(), expected, actual)
	}
}
