package bench

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"os"
)

var harJson = readHarContent()
var harObj = Unmarshal()

func readHarContent() []byte {
	content, err := os.ReadFile("vlang.io.har.pretty.json")
	if err != nil {
		log.Fatalf("decoding failed: %v", err)
	}
	return content
}

func Decode() *har {
	var h har
	err := json.NewDecoder(bytes.NewBuffer(harJson)).Decode(&h)
	if err != nil {
		log.Fatalf("decoding failed: %v", err)
	}
	return &h
}

func Unmarshal() *har {
	var h har
	err := json.Unmarshal(harJson, &h)
	if err != nil {
		log.Fatalf("decoding failed: %v", err)
	}
	return &h
}

func UnmarshalFromReader() *har {
	out, _ := io.ReadAll(bytes.NewBuffer(harJson))
	var h har
	err := json.Unmarshal(out, &h)
	if err != nil {
		log.Fatalf("decoding failed: %v", err)
	}
	return &h
}

func Encode() {
	err := json.NewEncoder(new(bytes.Buffer)).Encode(harObj)
	if err != nil {
		log.Fatalf("encoding failed: %v", err)
	}
}

func Marshal() {
	_, err := json.Marshal(harObj)
	if err != nil {
		log.Fatalf("decoding failed: %v", err)
	}
}

func MarshalToWriter() {
	out, err := json.Marshal(harObj)
	if err != nil {
		log.Fatalf("decoding failed: %v", err)
	}
	new(bytes.Buffer).Write(out)
}
