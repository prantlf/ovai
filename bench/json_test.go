package bench

import (
	"testing"
)

func BenchmarkDecode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Decode()
	}
}

func BenchmarkUnmarshal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Unmarshal()
	}
}

func BenchmarkUnmarshalFromReader(b *testing.B) {
	for i := 0; i < b.N; i++ {
		UnmarshalFromReader()
	}
}

func BenchmarkEncode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Encode()
	}
}

func BenchmarkMarshal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Marshal()
	}
}

func BenchmarkMarshalToWriter(b *testing.B) {
	for i := 0; i < b.N; i++ {
		MarshalToWriter()
	}
}
