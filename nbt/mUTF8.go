package nbt

import (
	"bytes"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

// This implementation correctly handles the special null character encoding and surrogate pairs.

// decodeMUTF8 decodes a byte slice from Java's Modified UTF-8 format into a Go string.
func decodeMUTF8(b []byte) string {
	var runes []rune
	for i := 0; i < len(b); {
		c := b[i]
		switch {
		// 1-byte sequence (ASCII)
		case c < 0x80:
			i++
			runes = append(runes, rune(c))

		// 2-byte sequence
		case c&0xE0 == 0xC0:
			if i+1 >= len(b) {
				runes = append(runes, unicode.ReplacementChar)
				i++ // Move past invalid byte
				continue
			}
			c2 := b[i+1]
			if c2&0xC0 != 0x80 {
				runes = append(runes, unicode.ReplacementChar)
				i++ // Move past invalid byte
				continue
			}
			// Check for special null character encoding
			if c == 0xC0 && c2 == 0x80 {
				runes = append(runes, 0)
			} else {
				runes = append(runes, (rune(c&0x1F)<<6)|rune(c2&0x3F))
			}
			i += 2

		// 3-byte sequence (or start of a surrogate pair)
		case c&0xF0 == 0xE0:
			if i+2 >= len(b) {
				runes = append(runes, unicode.ReplacementChar)
				i++ // Move past invalid byte
				continue
			}
			c2, c3 := b[i+1], b[i+2]
			if c2&0xC0 != 0x80 || c3&0xC0 != 0x80 {
				runes = append(runes, unicode.ReplacementChar)
				i++ // Move past invalid byte
				continue
			}
			r := (rune(c&0x0F) << 12) | (rune(c2&0x3F) << 6) | rune(c3&0x3F)

			// Check if it's a high surrogate, indicating a 6-byte sequence
			if r >= 0xD800 && r <= 0xDBFF {
				// We expect another 3-byte sequence for the low surrogate
				if i+5 >= len(b) {
					runes = append(runes, unicode.ReplacementChar)
					i += 3
					continue
				}
				c4, c5, c6 := b[i+3], b[i+4], b[i+5]
				if c4&0xF0 != 0xE0 || c5&0xC0 != 0x80 || c6&0xC0 != 0x80 {
					runes = append(runes, unicode.ReplacementChar)
					i += 3
					continue
				}
				lowSurrogate := (rune(c4&0x0F) << 12) | (rune(c5&0x3F) << 6) | rune(c6&0x3F)

				// Combine the high and low surrogates
				runes = append(runes, utf16.DecodeRune(r, lowSurrogate))
				i += 6
			} else {
				runes = append(runes, r)
				i += 3
			}

		// Invalid start byte
		default:
			runes = append(runes, unicode.ReplacementChar)
			i++
		}
	}
	return string(runes)
}

// encodeMUTF8 encodes a Go string into a byte slice in Java's Modified UTF-8 format.
func encodeMUTF8(s string) []byte {
	var buf bytes.Buffer
	for _, r := range s {
		switch {
		// Special case: Null character
		case r == 0:
			buf.WriteByte(0xC0)
			buf.WriteByte(0x80)

		// 1-byte sequence (ASCII)
		case r > 0 && r <= 0x7F:
			buf.WriteByte(byte(r))

		// 2-byte sequence
		case r <= 0x7FF:
			buf.WriteByte(byte(0xC0 | (r >> 6)))
			buf.WriteByte(byte(0x80 | (r & 0x3F)))

		// 3-byte sequence (for BMP characters, excluding surrogates)
		case r <= 0xFFFF && (r < 0xD800 || r > 0xDFFF):
			buf.WriteByte(byte(0xE0 | (r >> 12)))
			buf.WriteByte(byte(0x80 | (r >> 6 & 0x3F)))
			buf.WriteByte(byte(0x80 | (r & 0x3F)))

		// 6-byte sequence (for supplementary characters using surrogate pairs)
		case r > 0xFFFF:
			r1, r2 := utf16.EncodeRune(r)

			// Encode high surrogate
			buf.WriteByte(byte(0xE0 | (r1 >> 12)))
			buf.WriteByte(byte(0x80 | (r1 >> 6 & 0x3F)))
			buf.WriteByte(byte(0x80 | (r1 & 0x3F)))

			// Encode low surrogate
			buf.WriteByte(byte(0xE0 | (r2 >> 12)))
			buf.WriteByte(byte(0x80 | (r2 >> 6 & 0x3F)))
			buf.WriteByte(byte(0x80 | (r2 & 0x3F)))

		// Invalid rune (e.g., unpaired surrogate) - encode as replacement character
		default:
			replacement, _ := utf8.DecodeRuneInString(string(unicode.ReplacementChar))
			buf.WriteByte(byte(0xE0 | (replacement >> 12)))
			buf.WriteByte(byte(0x80 | (replacement >> 6 & 0x3F)))
			buf.WriteByte(byte(0x80 | (replacement & 0x3F)))
		}
	}
	return buf.Bytes()
}
