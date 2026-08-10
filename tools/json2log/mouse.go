// Mouse payload codec (5 bytes).
//
// A mouse report (port 1) is 5 bytes packed as:
//
//	byte 0: button      (bit 0 = left, bit 1 = right, bit 2 = middle)
//	byte 1: x_lo        (low 8 bits of x)
//	byte 2: x_hi4 | y_lo4 << 4
//	byte 3: y_hi8
//	byte 4: wheel       (signed 8-bit)
//
// x and y are 12-bit signed values (range -2048..2047). See
// player/main/applog_good/hidcodec.c (M10BuffEncode / M10BuffDecode)
// for the authoritative C implementation.
package main

import "fmt"

// Mouse button bits.
const (
	MouseButtonLeft   = 0x01
	MouseButtonRight  = 0x02
	MouseButtonMiddle = 0x04
)

// MouseReport is the decoded form of a 5-byte mouse record payload.
type MouseReport struct {
	Button uint8 // bit 0=L, 1=R, 2=M
	X      int16 // -2048..2047
	Y      int16 // -2048..2047
	Wheel  int8  // -128..127
}

// MouseEncode packs the report into 5 bytes.
func MouseEncode(button uint8, x, y int16, wheel int8) [5]byte {
	var out [5]byte
	out[0] = button

	// Mask to 12 bits: this also normalises negatives into the unsigned
	// 12-bit two's-complement representation.
	ux := uint16(uint16(int16(x)) & 0x0FFF)
	uy := uint16(uint16(int16(y)) & 0x0FFF)

	out[1] = byte(ux & 0xFF)
	out[2] = byte((ux>>8)&0x0F) | byte((uy&0x0F)<<4)
	out[3] = byte((uy >> 4) & 0xFF)
	out[4] = byte(int8(wheel))
	return out
}

// MouseDecode unpacks a 5-byte payload.
func MouseDecode(b [5]byte) MouseReport {
	button := b[0]
	wheel := int8(b[4])

	ux := uint16(b[1]) | (uint16(b[2]&0x0F) << 8)
	uy := (uint16(b[2]) >> 4) | (uint16(b[3]) << 4)

	x := uint16(ux & 0x0FFF)
	if x&0x0800 != 0 {
		x |= 0xF000
	}
	y := uint16(uy & 0x0FFF)
	if y&0x0800 != 0 {
		y |= 0xF000
	}
	return MouseReport{Button: button, X: int16(x), Y: int16(y), Wheel: wheel}
}

// String formats the report for logging.
func (r MouseReport) String() string {
	return fmt.Sprintf("Mouse{btn=0x%02x x=%d y=%d w=%d}", r.Button, r.X, r.Y, r.Wheel)
}
