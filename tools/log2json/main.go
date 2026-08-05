// log2json - bidirectional .log <-> .json converter.
//
// Usage:
//
//	log2json -f <input> [-o <output>]
//
// Direction is decided by the input file's extension:
//
//	.log  -> .json
//	.json -> .log
//
// If -o is omitted the output is the input path with the extension swapped.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "log2json:", err)
		os.Exit(1)
	}
}

func run() error {
	inPath := flag.String("f", "", "input file (.log or .json)")
	outPath := flag.String("o", "", "output file (default: same name with opposite extension)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: log2json -f <input> [-o <output>]\n\n")
		fmt.Fprintf(os.Stderr, "Convert between binary applog (.log) and JSON action envelope (.json).\n")
		fmt.Fprintf(os.Stderr, "The direction is decided by the input file extension.\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *inPath == "" {
		flag.Usage()
		return fmt.Errorf("missing -f")
	}

	dir := direction(*inPath)
	if dir == dirUnknown {
		return fmt.Errorf("unrecognized input extension on %q (expected .log or .json)", *inPath)
	}

	out := *outPath
	if out == "" {
		out = swapExt(*inPath, dir.targetExt())
	} else if !strings.HasSuffix(strings.ToLower(out), dir.targetExt()) {
		// explicit -o without an extension: append the canonical one
		out = out + dir.targetExt()
	}

	// ensure parent dir of out exists
	if d := filepath.Dir(out); d != "" && d != "." {
		_ = os.MkdirAll(d, 0o755)
	}

	switch dir {
	case dirLogToJSON:
		return convertLogToJSON(*inPath, out)
	case dirJSONToLog:
		return convertJSONToLog(*inPath, out)
	}
	return fmt.Errorf("internal: unhandled direction")
}

type directionT int

const (
	dirUnknown directionT = iota
	dirLogToJSON
	dirJSONToLog
)

func (d directionT) targetExt() string {
	switch d {
	case dirLogToJSON:
		return ".json"
	case dirJSONToLog:
		return ".log"
	}
	return ""
}

// direction picks the conversion direction from the input extension
// (case-insensitive).
func direction(path string) directionT {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".log":
		return dirLogToJSON
	case ".json":
		return dirJSONToLog
	}
	return dirUnknown
}

// swapExt returns path with the extension replaced.
func swapExt(path, newExt string) string {
	base := path[:len(path)-len(filepath.Ext(path))]
	return base + newExt
}

func convertLogToJSON(in, out string) error {
	hdr, recs, err := ReadLogFile(in)
	if err != nil {
		return err
	}
	env := LogToJSON(hdr, recs)
	if err := WriteJSONFile(out, env); err != nil {
		return err
	}
	fmt.Printf("log -> json: %s (%d events)\n", out, len(env.Events))
	return nil
}

func convertJSONToLog(in, out string) error {
	env, err := ReadJSONFile(in)
	if err != nil {
		return err
	}
	hdr, recs, hasMouse, hasKbd := JSONToLog(env)
	if err := WriteLogFile(out, hdr, recs, hasMouse, hasKbd); err != nil {
		return err
	}
	fmt.Printf("json -> log: %s (version=%s, mouse=%d, kbd=%d)\n",
		out, hdr.Version, boolToInt(hasMouse), boolToInt(hasKbd))
	return nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
