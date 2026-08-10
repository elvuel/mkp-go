// json2log - convert adumj/ajson2log JSON back to binary .log.
//
// Usage:
//
//	json2log -f <input.json> [-o <output.log>]
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
		fmt.Fprintln(os.Stderr, "json2log:", err)
		os.Exit(1)
	}
}

func run() error {
	inPath := flag.String("f", "", "input JSON file (.json)")
	outPath := flag.String("o", "", "output log file (default: same name with .log extension)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: json2log -f <input.json> [-o <output.log>]\n\n")
		fmt.Fprintf(os.Stderr, "Convert firmware action JSON (.json) to binary applog (.log).\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *inPath == "" {
		flag.Usage()
		return fmt.Errorf("missing -f")
	}
	if strings.ToLower(filepath.Ext(*inPath)) != ".json" {
		return fmt.Errorf("unrecognized input extension on %q (expected .json)", *inPath)
	}

	out := *outPath
	if out == "" {
		out = swapExt(*inPath, ".log")
	} else if filepath.Ext(out) == "" {
		// explicit -o without an extension: append the canonical one
		out = out + ".log"
	}

	// ensure parent dir of out exists
	if d := filepath.Dir(out); d != "" && d != "." {
		_ = os.MkdirAll(d, 0o755)
	}

	return convertJSONToLog(*inPath, out)
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
	hdr, recs, hasMouse, hasKbd, err := JSONToLog(env)
	if err != nil {
		return err
	}
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
