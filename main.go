// Command urlcode is used to escape and unescape URL query and path strings.
// It is just a simple wrapper around url.QueryEscape and Unescape and equivalent path functions.
package main

import (
	"flag"
	"io"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"unicode"
)

type Codec interface {
	Encode(string) (string, error)
	Decode(string) (string, error)
}

type queryCodec struct{}

func (queryCodec) Encode(s string) (string, error) {
	return url.QueryEscape(s), nil
}

func (queryCodec) Decode(s string) (string, error) {
	return url.QueryUnescape(s)
}

type pathCodec struct{}

func (pathCodec) Encode(s string) (string, error) {
	return url.PathEscape(s), nil
}

func (pathCodec) Decode(s string) (string, error) {
	return url.PathUnescape(s)
}

func main() {
	opDecode := flag.Bool("d", false, "Decode URL strings")
	usersep := flag.String("s", "", "Output separator (defaults to space)")
	pathCoding := flag.Bool("p", false, "Encode/decode path strings")
	flag.Parse()

	// Select codec (query or path)
	codec := Codec(queryCodec{})
	if *pathCoding {
		codec = pathCodec{}
	}

	// Configure coding op
	op := "encode"
	code := codec.Encode
	sep := "+"
	if *opDecode {
		op = "decode"
		code = codec.Decode
		sep = " "
	}

	// Parse user separator
	if *usersep != "" {
		sep = *usersep
		if uq, err := strconv.Unquote(`"` + sep + `"`); err == nil {
			sep = uq
		} else if sep == "\\0" {
			sep = "\x00"
		}
	}

	// Encode/decode each argument separately
	args := append([]string(nil), flag.Args()...)
	var err error
	for i, s := range args {
		args[i], err = code(s)
		if err != nil {
			log.Fatalf("unable to %s %q: %v", op, s, err)
		}
	}

	// Write output
	out := strings.Join(args, sep)
	io.WriteString(os.Stdout, out)

	// Check if we should print a newline after everything
	isPrint := true
	for _, r := range sep {
		isPrint = isPrint && (unicode.IsSpace(r) || unicode.IsPrint(r))
	}

	if isPrint && isTTY() {
		io.WriteString(os.Stdout, "\n")
	}
}

// isTTY attempts to determine whether the current stdout refers to a terminal.
func isTTY() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false // Assume it's not a TTY
	}
	return (fi.Mode() & os.ModeNamedPipe) != os.ModeNamedPipe
}
