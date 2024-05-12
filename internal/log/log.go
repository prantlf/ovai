package log

import (
	out "log"
	"os"
	"strings"
)

var IsDbg bool
var IsSrv bool
var IsNet bool

var _ = initLog()

func initLog() bool {
	out.SetFlags(out.Ldate | out.Lmicroseconds)
	cfg := os.Getenv("DEBUG")
	for _, part := range strings.Split(cfg, ",") {
		switch strings.ToLower(part) {
		case "ovai":
			IsDbg = true
		case "ovai*":
			IsDbg = true
			IsSrv = true
			IsNet = true
		case "ovai:srv":
			IsSrv = true
		case "ovai:net":
			IsNet = true
		case "ovai:*":
			IsSrv = true
			IsNet = true
		}
	}
	return IsDbg || IsSrv || IsNet
}

func Dbg(format string, args ...any) {
	if IsDbg {
		out.Printf(format, args...)
	}
}

func Srv(format string, args ...any) {
	if IsSrv {
		out.Printf(format, args...)
	}
}

func Net(format string, args ...any) {
	if IsNet {
		out.Printf(format, args...)
	}
}

func Ftl(format string, args ...any) {
	out.Fatalf(format, args...)
}
