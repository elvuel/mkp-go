// log2json - compatibility wrapper around tools/converter.
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

	converter "redux-solutions.com/mousekeypadplayer/tools/converter/lib"
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
		fmt.Fprintf(os.Stderr, "Compatibility wrapper around tools/converter.\n")
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

	dir := converter.DirectionFromPath(*inPath)
	if dir == converter.DirUnknown {
		return fmt.Errorf("unrecognized input extension on %q (expected .log or .json)", *inPath)
	}

	out := *outPath
	if out == "" {
		out = swapExt(*inPath, dir.TargetExt())
	} else if !strings.HasSuffix(strings.ToLower(out), dir.TargetExt()) {
		// explicit -o without an extension: append the canonical one
		out = out + dir.TargetExt()
	}

	// ensure parent dir of out exists
	if d := filepath.Dir(out); d != "" && d != "." {
		_ = os.MkdirAll(d, 0o755)
	}

	switch dir {
	case converter.DirLogToJSON:
		return converter.ConvertLogFileToJSONFile(*inPath, out)
	case converter.DirJSONToLog:
		return converter.ConvertJSONFileToLogFile(*inPath, out)
	}
	return fmt.Errorf("internal: unhandled direction")
}

// swapExt returns path with the extension replaced.
func swapExt(path, newExt string) string {
	base := path[:len(path)-len(filepath.Ext(path))]
	return base + newExt
}
