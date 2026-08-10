package converter

import (
	"fmt"
	"path/filepath"
	"strings"
)

type Direction int

const (
	DirUnknown Direction = iota
	DirLogToJSON
	DirJSONToLog
)

func (d Direction) TargetExt() string {
	switch d {
	case DirLogToJSON:
		return ".json"
	case DirJSONToLog:
		return ".log"
	}
	return ""
}

func ConvertLogFileToJSONFile(in, out string) error {
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

func ConvertJSONFileToLogFile(in, out string) error {
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

func convertLogToJSON(in, out string) error {
	return ConvertLogFileToJSONFile(in, out)
}

func convertJSONToLog(in, out string) error {
	return ConvertJSONFileToLogFile(in, out)
}

func DirectionFromPath(path string) Direction {
	switch ext := lowerExt(path); ext {
	case ".log":
		return DirLogToJSON
	case ".json":
		return DirJSONToLog
	}
	return DirUnknown
}

func lowerExt(path string) string {
	return strings.ToLower(filepath.Ext(path))
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
