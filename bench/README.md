# Benchmarks

Marshalling and unmarshalling is faster than encoding and decoding.

    ❯ go test -bench=.

    goos: darwin
    goarch: amd64
    pkg: github.com/prantlf/ovai/bench
    cpu: Intel(R) Core(TM) i7-8850H CPU @ 2.60GHz
    BenchmarkDecode-12                 	      52	  22209908 ns/op	21126355 B/op	    2176 allocs/op
    BenchmarkUnmarshal-12              	      58	  19268460 ns/op	 4350828 B/op	    2157 allocs/op
    BenchmarkUnmarshalFromReader-12    	      54	  22253316 ns/op	25459898 B/op	    2191 allocs/op
    BenchmarkEncode-12                 	     247	   4900571 ns/op	 4276109 B/op	       2 allocs/op
    BenchmarkMarshal-12                	     254	   4599008 ns/op	 4153713 B/op	       1 allocs/op
    BenchmarkMarshalToWriter-12        	     199	   6161806 ns/op	10433388 B/op	      10 allocs/op
    PASS
    ok  	github.com/prantlf/ovai/bench	10.126s
