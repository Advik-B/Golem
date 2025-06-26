package nbt

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// Move the regex compilation to the package level.
var numberRegex = regexp.MustCompile(`^[-+]?([0-9]+(\.[0-9]*)?|\.[0-9]+)([eE][-+]?[0-9]+)?[bBsSlLdDfF]?`)

// parser holds the state of the SNBT parsing.
type parser struct {
	s      string
	cursor int
}

// ParseSNBT converts an SNBT string into a Tag. It expects a CompoundTag.
func ParseSNBT(data string) (*CompoundTag, error) {
	p := &parser{s: data}
	val, err := p.parseValue()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	if p.hasMore() {
		return nil, p.error("trailing data found after top-level tag")
	}

	if compound, ok := val.(*CompoundTag); ok {
		return compound, nil
	}
	return nil, p.error("expected compound tag at top level")
}

func (p *parser) error(msg string) error {
	return fmt.Errorf("%s at position %d", msg, p.cursor)
}

func (p *parser) skipWhitespace() {
	for p.hasMore() && unicode.IsSpace(rune(p.s[p.cursor])) {
		p.cursor++
	}
}

func (p *parser) hasMore() bool {
	return p.cursor < len(p.s)
}

func (p *parser) peek() rune {
	if !p.hasMore() {
		return 0
	}
	return rune(p.s[p.cursor])
}

func (p *parser) consume() rune {
	r := p.peek()
	p.cursor++
	return r
}

func (p *parser) expect(r rune) error {
	p.skipWhitespace()
	if p.peek() != r {
		return p.error(fmt.Sprintf("expected '%c'", r))
	}
	p.cursor++
	return nil
}

func (p *parser) parseValue() (Tag, error) {
	p.skipWhitespace()
	if !p.hasMore() {
		return nil, p.error("expected value")
	}
	switch p.peek() {
	case '{':
		return p.parseCompound()
	case '[':
		return p.parseListOrArray()
	case '"', '\'':
		s, err := p.parseQuotedString()
		if err != nil {
			return nil, err
		}
		return &StringTag{Value: s}, nil
	default:
		// An unquoted value must be a number or a boolean constant.
		return p.parseNumberOrBoolean()
	}
}

func (p *parser) parseCompound() (*CompoundTag, error) {
	if err := p.expect('{'); err != nil {
		return nil, err
	}
	tag := NewCompoundTag()
	p.skipWhitespace()

	if p.peek() == '}' { // Handle empty compound
		p.cursor++
		return tag, nil
	}

	for {
		key, err := p.parseKey()
		if err != nil {
			return nil, err
		}

		if err := p.expect(':'); err != nil {
			return nil, err
		}
		val, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		tag.Put(key, val)

		p.skipWhitespace()
		if p.peek() != ',' {
			break // No comma, must be the end of the compound
		}
		p.cursor++ // Consume the comma

		p.skipWhitespace()
		if p.peek() == '}' {
			return nil, p.error("trailing comma in compound tag")
		}
	}

	if err := p.expect('}'); err != nil {
		return nil, err
	}
	return tag, nil
}

func (p *parser) parseKey() (string, error) {
	p.skipWhitespace()
	if p.peek() == '"' || p.peek() == '\'' {
		return p.parseQuotedString()
	}
	return p.parseUnquotedString()
}

func (p *parser) parseListOrArray() (Tag, error) {
	if err := p.expect('['); err != nil {
		return nil, err
	}
	p.skipWhitespace()

	// Check for special array format like [B;...]
	if p.cursor+1 < len(p.s) && p.s[p.cursor+1] == ';' {
		switch p.s[p.cursor] {
		case 'B':
			return p.parseByteArray()
		case 'I':
			return p.parseIntArray()
		case 'L':
			return p.parseLongArray()
		}
	}

	return p.parseList()
}

func (p *parser) parseList() (*ListTag, error) {
	list := &ListTag{}
	p.skipWhitespace()
	if p.peek() == ']' {
		p.cursor++
		return list, nil
	}

	for {
		val, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		if err := list.Add(val); err != nil {
			return nil, p.error(err.Error())
		}
		p.skipWhitespace()
		if p.peek() != ',' {
			break
		}
		p.cursor++

		p.skipWhitespace()
		if p.peek() == ']' {
			return nil, p.error("trailing comma in list tag")
		}
	}

	if err := p.expect(']'); err != nil {
		return nil, err
	}
	return list, nil
}

func (p *parser) parseByteArray() (*ByteArrayTag, error) {
	p.cursor += 2 // Skip 'B;'
	var values []int8
	p.skipWhitespace()
	if p.peek() == ']' {
		p.cursor++
		return &ByteArrayTag{Value: values}, nil
	}
	for {
		v, err := p.parseNumber(TagByte)
		if err != nil {
			return nil, err
		}
		values = append(values, v.(*ByteTag).Value)
		p.skipWhitespace()
		if p.peek() != ',' {
			break
		}
		p.cursor++
		p.skipWhitespace()
		if p.peek() == ']' {
			return nil, p.error("trailing comma in byte array tag")
		}
	}
	if err := p.expect(']'); err != nil {
		return nil, err
	}
	return &ByteArrayTag{Value: values}, nil
}

func (p *parser) parseIntArray() (*IntArrayTag, error) {
	p.cursor += 2 // Skip 'I;'
	var values []int32
	p.skipWhitespace()
	if p.peek() == ']' {
		p.cursor++
		return &IntArrayTag{Value: values}, nil
	}
	for {
		v, err := p.parseNumber(TagInt)
		if err != nil {
			return nil, err
		}
		values = append(values, v.(*IntTag).Value)
		p.skipWhitespace()
		if p.peek() != ',' {
			break
		}
		p.cursor++
		p.skipWhitespace()
		if p.peek() == ']' {
			return nil, p.error("trailing comma in int array tag")
		}
	}
	if err := p.expect(']'); err != nil {
		return nil, err
	}
	return &IntArrayTag{Value: values}, nil
}

func (p *parser) parseLongArray() (*LongArrayTag, error) {
	p.cursor += 2 // Skip 'L;'
	var values []int64
	p.skipWhitespace()
	if p.peek() == ']' {
		p.cursor++
		return &LongArrayTag{Value: values}, nil
	}
	for {
		v, err := p.parseNumber(TagLong)
		if err != nil {
			return nil, err
		}
		values = append(values, v.(*LongTag).Value)
		p.skipWhitespace()
		if p.peek() != ',' {
			break
		}
		p.cursor++
		p.skipWhitespace()
		if p.peek() == ']' {
			return nil, p.error("trailing comma in long array tag")
		}
	}
	if err := p.expect(']'); err != nil {
		return nil, err
	}
	return &LongArrayTag{Value: values}, nil
}

var unquotedStringRegex = regexp.MustCompile(`^[a-zA-Z0-9_.-]+`)

func (p *parser) parseUnquotedString() (string, error) {
	p.skipWhitespace()
	match := unquotedStringRegex.FindString(p.s[p.cursor:])
	if match == "" {
		return "", p.error("expected unquoted string")
	}
	p.cursor += len(match)
	return match, nil
}

func (p *parser) parseQuotedString() (string, error) {
	p.skipWhitespace()
	quote := p.consume()
	if quote != '"' && quote != '\'' {
		return "", p.error("expected quote")
	}

	var sb strings.Builder
	escaped := false
	for p.hasMore() {
		c := p.consume()
		if escaped {
			sb.WriteRune(c)
			escaped = false
		} else if c == '\\' {
			escaped = true
		} else if c == quote {
			return sb.String(), nil
		} else {
			sb.WriteRune(c)
		}
	}
	return "", p.error("unterminated string")
}

func (p *parser) parseNumberOrBoolean() (Tag, error) {
	p.skipWhitespace()
	remaining := p.s[p.cursor:]

	// Check for boolean constants before attempting to parse a number
	if len(remaining) >= 4 && remaining[:4] == "true" {
		p.cursor += 4
		return &ByteTag{Value: 1}, nil
	}
	if len(remaining) >= 5 && remaining[:5] == "false" {
		p.cursor += 5
		return &ByteTag{Value: 0}, nil
	}

	// Not a boolean, so it must be a number
	tag, err := p.parseNumber(0)
	if err != nil {
		// If it's not a boolean and not a number, it's an invalid value.
		return nil, p.error("expected value, found invalid unquoted string")
	}
	return tag, nil
}

// parseNumber parses a number string, respecting an optional suffix and a contextual type.
func (p *parser) parseNumber(contextualType TagID) (Tag, error) {
	p.skipWhitespace()

	// A simple regex to find a number-like pattern.
	// This is more robust than the previous implementation.
	match := numberRegex.FindString(p.s[p.cursor:])
	if match == "" {
		return nil, p.error("expected a number")
	}

	p.cursor += len(match)
	numStr := match
	suffix := rune(0)

	lastChar := rune(numStr[len(numStr)-1])
	if !unicode.IsDigit(lastChar) && lastChar != '.' {
		suffix = unicode.ToLower(lastChar)
		numStr = numStr[:len(numStr)-1]
	}

	// Explicit suffix always wins
	if suffix != 0 {
		switch suffix {
		case 'b':
			v, err := strconv.ParseInt(numStr, 10, 8)
			return &ByteTag{Value: int8(v)}, err
		case 's':
			v, err := strconv.ParseInt(numStr, 10, 16)
			return &ShortTag{Value: int16(v)}, err
		case 'l':
			v, err := strconv.ParseInt(numStr, 10, 64)
			return &LongTag{Value: v}, err
		case 'f':
			v, err := strconv.ParseFloat(numStr, 32)
			return &FloatTag{Value: float32(v)}, err
		case 'd':
			v, err := strconv.ParseFloat(numStr, 64)
			return &DoubleTag{Value: v}, err
		}
	}

	// No suffix, use context or infer
	if strings.ContainsAny(numStr, ".eE") {
		v, err := strconv.ParseFloat(numStr, 64)
		if contextualType == TagFloat {
			return &FloatTag{Value: float32(v)}, err
		}
		return &DoubleTag{Value: v}, err
	}

	// Integer without suffix, use context
	switch contextualType {
	case TagByte:
		v, err := strconv.ParseInt(numStr, 10, 8)
		return &ByteTag{Value: int8(v)}, err
	case TagShort:
		v, err := strconv.ParseInt(numStr, 10, 16)
		return &ShortTag{Value: int16(v)}, err
	case TagLong:
		v, err := strconv.ParseInt(numStr, 10, 64)
		return &LongTag{Value: v}, err
	default: // Default to IntTag if no context or unknown
		v, err := strconv.ParseInt(numStr, 10, 32)
		return &IntTag{Value: int32(v)}, err
	}
}
