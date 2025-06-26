package nbt

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Printer converts an NBT Tag to its SNBT string representation.
type Printer struct {
	pretty      bool   // If true, adds indentation and newlines.
	indentation string // The string to use for one level of indentation.
	depth       int
	builder     strings.Builder
}

// NewPrettyPrinter creates a new SNBT printer for formatted, human-readable output.
func NewPrettyPrinter(indent string) *Printer {
	return &Printer{pretty: true, indentation: indent}
}

// NewCompactPrinter creates a new SNBT printer for compact, single-line output.
func NewCompactPrinter() *Printer {
	return &Printer{pretty: false, indentation: ""}
}

// Print converts the tag to an SNBT string using the printer's mode.
func (p *Printer) Print(tag Tag) string {
	p.builder.Reset()
	p.depth = 0
	p.writeTag(tag)
	return p.builder.String()
}

func (p *Printer) writeTag(tag Tag) {
	switch v := tag.(type) {
	// ... (cases for primitive tags are unchanged) ...
	case *EndTag:
	case *ByteTag:
		p.builder.WriteString(fmt.Sprintf("%db", v.Value))
	case *ShortTag:
		p.builder.WriteString(fmt.Sprintf("%ds", v.Value))
	case *IntTag:
		p.builder.WriteString(fmt.Sprintf("%d", v.Value))
	case *LongTag:
		p.builder.WriteString(fmt.Sprintf("%dL", v.Value))
	case *FloatTag:
		p.builder.WriteString(fmt.Sprintf("%gf", v.Value))
	case *DoubleTag:
		p.builder.WriteString(fmt.Sprintf("%gd", v.Value))
	case *StringTag:
		p.builder.WriteString(quoteAndEscape(v.Value))
	case *ByteArrayTag:
		p.builder.WriteString("[B;")
		for i, b := range v.Value {
			if i > 0 {
				p.builder.WriteString(",")
			}
			p.builder.WriteString(fmt.Sprintf(" %db", b))
		}
		p.builder.WriteString("]")
	case *IntArrayTag:
		p.builder.WriteString("[I;")
		for i, i32 := range v.Value {
			if i > 0 {
				p.builder.WriteString(",")
			}
			p.builder.WriteString(fmt.Sprintf(" %d", i32))
		}
		p.builder.WriteString("]")
	case *LongArrayTag:
		p.builder.WriteString("[L;")
		for i, i64 := range v.Value {
			if i > 0 {
				p.builder.WriteString(",")
			}
			p.builder.WriteString(fmt.Sprintf(" %dL", i64))
		}
		p.builder.WriteString("]")
	case *ListTag:
		p.writeList(v)
	case *CompoundTag:
		p.writeCompound(v)
	}
}

func (p *Printer) writeList(t *ListTag) {
	if len(t.Value) == 0 {
		p.builder.WriteString("[]")
		return
	}

	p.builder.WriteString("[")
	if p.pretty {
		p.depth++
	}

	for i, tag := range t.Value {
		if i > 0 {
			p.builder.WriteString(",")
			if !p.pretty {
				p.builder.WriteString(" ")
			}
		}

		if p.pretty {
			p.builder.WriteString("\n")
			p.builder.WriteString(strings.Repeat(p.indentation, p.depth))
		}

		p.writeTag(tag)
	}

	if p.pretty {
		p.depth--
		p.builder.WriteString("\n")
		p.builder.WriteString(strings.Repeat(p.indentation, p.depth))
	}
	p.builder.WriteString("]")
}

var keyPattern = regexp.MustCompile(`^[A-Za-z0-9._+-]+$`)

func handleKey(key string) string {
	if keyPattern.MatchString(key) {
		return key
	}
	return quoteAndEscape(key)
}

func (p *Printer) writeCompound(t *CompoundTag) {
	if len(t.Value) == 0 {
		p.builder.WriteString("{}")
		return
	}

	p.builder.WriteString("{")
	if p.pretty {
		p.depth++
	}

	keys := make([]string, 0, len(t.Value))
	for k := range t.Value {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for i, key := range keys {
		if i > 0 {
			p.builder.WriteString(",")
		}

		if p.pretty {
			p.builder.WriteString("\n")
			p.builder.WriteString(strings.Repeat(p.indentation, p.depth))
		} else if i > 0 {
			p.builder.WriteString(" ")
		}

		p.builder.WriteString(handleKey(key))
		p.builder.WriteString(": ")
		p.writeTag(t.Value[key])
	}

	if p.pretty {
		p.depth--
		p.builder.WriteString("\n")
		p.builder.WriteString(strings.Repeat(p.indentation, p.depth))
	}
	p.builder.WriteString("}")
}

// ToPrettySNBT converts a tag to its pretty-printed SNBT representation.
func ToPrettySNBT(tag Tag) string {
	return NewPrettyPrinter("  ").Print(tag)
}

// ToCompactSNBT converts a tag to its compact SNBT representation.
func ToCompactSNBT(tag Tag) string {
	return NewCompactPrinter().Print(tag)
}
