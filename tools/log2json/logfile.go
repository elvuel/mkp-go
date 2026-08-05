// Binary applog (.log) reader and writer.
//
// File layout (see docs/log-file-format.md):
//
//	V1: "MKv1" (4 bytes)  + records...
//	V2: "MKv2" (4 bytes)  + meta(12)  + records...
//
//	record: portnum(1) + delta(4 LE) + length(2 LE) + payload(length)
//
// Mouse (port 1) payloads are 5-byte mouse reports (see mouse.go).
// Keyboard (port 2) payloads are HID Boot Keyboard reports (8 bytes).
package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

// Magic strings.
const (
	MagicV1 = "MKv1"
	MagicV2 = "MKv2"

	MagicLength      = 4
	MetaLength       = 12
	HeaderLengthV1   = MagicLength
	HeaderLengthV2   = MagicLength + MetaLength
	RecordHeaderSize = 1 + 4 + 2 // portnum + delta + length
	MaxPayloadLength = 64        // sanity cap; firmware maxes out at 8
)

// Meta is the V2 file header metadata block.
type Meta struct {
	Width  uint16
	Height uint16
	StartX uint16 // stored unsigned; treated as a position
	StartY uint16
	// 4 bytes reserved (always zero)
}

// LogHeader is the parsed V1/V2 file header.
type LogHeader struct {
	Version string // "MKv1" or "MKv2"
	Meta    Meta   // zero-valued for V1
}

// Record is one decoded record from the log stream.
type Record struct {
	Port    uint8
	Delta   int32 // milliseconds since previous record
	Length  uint16
	Payload []byte
}

// HeaderSize returns the byte size of the file header for this version.
func (h LogHeader) HeaderSize() int {
	if h.Version == MagicV2 {
		return HeaderLengthV2
	}
	return HeaderLengthV1
}

// ReadLogFile reads an entire log file into memory and returns the
// header and a slice of records.
func ReadLogFile(path string) (LogHeader, []Record, error) {
	f, err := os.Open(path)
	if err != nil {
		return LogHeader{}, nil, err
	}
	defer f.Close()

	hdr, err := readLogHeader(f)
	if err != nil {
		return LogHeader{}, nil, err
	}

	var recs []Record
	for {
		rec, err := readRecord(f)
		if err == io.EOF {
			break
		}
		if err != nil {
			return hdr, nil, err
		}
		recs = append(recs, rec)
	}
	return hdr, recs, nil
}

func readLogHeader(r io.Reader) (LogHeader, error) {
	var magic [MagicLength]byte
	if _, err := io.ReadFull(r, magic[:]); err != nil {
		if err == io.EOF {
			return LogHeader{}, errors.New("empty log file")
		}
		return LogHeader{}, fmt.Errorf("read magic: %w", err)
	}
	switch string(magic[:]) {
	case MagicV1:
		return LogHeader{Version: MagicV1}, nil
	case MagicV2:
		var meta Meta
		if err := binary.Read(r, binary.LittleEndian, &meta); err != nil {
			return LogHeader{}, fmt.Errorf("read meta: %w", err)
		}
		// Skip the 4 reserved bytes that follow the meta struct so the
		// record stream starts at the correct offset.
		var reserved uint32
		if err := binary.Read(r, binary.LittleEndian, &reserved); err != nil {
			return LogHeader{}, fmt.Errorf("read reserved bytes: %w", err)
		}
		return LogHeader{Version: MagicV2, Meta: meta}, nil
	default:
		return LogHeader{}, fmt.Errorf("bad magic %q", string(magic[:]))
	}
}

func readRecord(r io.Reader) (Record, error) {
	var hdr [RecordHeaderSize]byte
	if _, err := io.ReadFull(r, hdr[:]); err != nil {
		return Record{}, err
	}
	rec := Record{
		Port:   hdr[0],
		Delta:  int32(binary.LittleEndian.Uint32(hdr[1:5])),
		Length: binary.LittleEndian.Uint16(hdr[5:7]),
	}
	if rec.Length > MaxPayloadLength {
		return rec, fmt.Errorf("record length %d exceeds cap %d", rec.Length, MaxPayloadLength)
	}
	if rec.Length > 0 {
		buf := make([]byte, rec.Length)
		if _, err := io.ReadFull(r, buf); err != nil {
			return rec, fmt.Errorf("read payload: %w", err)
		}
		rec.Payload = buf
	}
	return rec, nil
}

// WriteLogFile writes records to path with the given header. A
// stop trailer is appended automatically if the corresponding
// hasMouse / hasKeyboard flag is set (matching the C ajson2log
// behaviour: 3 zero-payload mouse records, 1 zero-payload keyboard
// record, each with delta=8).
func WriteLogFile(path string, hdr LogHeader, recs []Record, hasMouse, hasKeyboard bool) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := writeLogHeader(f, hdr); err != nil {
		return err
	}
	for _, rec := range recs {
		if err := writeRecord(f, rec); err != nil {
			return err
		}
	}
	writeStopTrailer(f, hasMouse, hasKeyboard)
	return nil
}

func writeLogHeader(w io.Writer, hdr LogHeader) error {
	if hdr.Version == MagicV2 {
		if _, err := w.Write([]byte(MagicV2)); err != nil {
			return err
		}
		if err := binary.Write(w, binary.LittleEndian, hdr.Meta); err != nil {
			return err
		}
		// 4 reserved zero bytes
		var zero uint32
		if err := binary.Write(w, binary.LittleEndian, &zero); err != nil {
			return err
		}
	} else {
		if _, err := w.Write([]byte(MagicV1)); err != nil {
			return err
		}
	}
	return nil
}

func writeRecord(w io.Writer, rec Record) error {
	var hdr [RecordHeaderSize]byte
	hdr[0] = rec.Port
	binary.LittleEndian.PutUint32(hdr[1:5], uint32(rec.Delta))
	binary.LittleEndian.PutUint16(hdr[5:7], rec.Length)
	if _, err := w.Write(hdr[:]); err != nil {
		return err
	}
	if rec.Length > 0 {
		if _, err := w.Write(rec.Payload); err != nil {
			return err
		}
	}
	return nil
}

func writeStopTrailer(w io.Writer, hasMouse, hasKeyboard bool) {
	const delta = int32(8)
	if hasMouse {
		for i := 0; i < 3; i++ {
			_ = writeRecord(w, Record{Port: 1, Delta: delta, Length: 5, Payload: make([]byte, 5)})
		}
	}
	if hasKeyboard {
		_ = writeRecord(w, Record{Port: 2, Delta: delta, Length: 8, Payload: make([]byte, 8)})
	}
}
