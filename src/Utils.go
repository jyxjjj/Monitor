package main

import (
	"strings"
)

func removeAllRightZeroAndPointForFloatString(f string) string {
	for strings.HasSuffix(f, "0") {
		f = f[:len(f)-1]
	}
	if strings.HasSuffix(f, ".") {
		f = f[:len(f)-1]
	}
	return f
}
