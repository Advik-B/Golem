package nbt

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
)

// Read reads a single named tag from the given reader.
// The NBT format must not be compressed.
func Read(r io.Reader) (NamedTag, error) {
	var namedTag NamedTag
	br := bufio.NewReader(r)

	id, err := br.ReadByte()
	if err != nil {
		if err == io.EOF {
			return namedTag, fmt.Errorf("unexpected EOF while reading tag ID")
		}
		return namedTag, err
	}

	if id == TagEnd {
		namedTag.Tag = &EndTag{}
		return namedTag, nil
	}

	namedTag.Name, err = readUTF(br)
	if err != nil {
		return namedTag, err
	}

	tag, err := newTag(id)
	if err != nil {
		return namedTag, err
	}

	if err := tag.read(br); err != nil {
		return namedTag, err
	}
	namedTag.Tag = tag

	return namedTag, nil
}

// Write writes a single named tag to the given writer.
// The data will not be compressed.
func Write(w io.Writer, nbt NamedTag) error {
	bw := bufio.NewWriter(w)

	if err := bw.WriteByte(nbt.Tag.ID()); err != nil {
		return err
	}

	if nbt.Tag.ID() != TagEnd {
		if err := writeUTF(bw, nbt.Name); err != nil {
			return err
		}
		if err := nbt.Tag.write(bw); err != nil {
			return err
		}
	}

	return bw.Flush()
}

// ReadCompressed reads a single named tag from a GZIP-compressed reader.
func ReadCompressed(r io.Reader) (NamedTag, error) {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return NamedTag{}, err
	}
	defer gzr.Close()

	return Read(gzr)
}

// WriteCompressed writes a single named tag to a writer, compressing it with GZIP.
func WriteCompressed(w io.Writer, nbt NamedTag) error {
	gzw := gzip.NewWriter(w)
	defer gzw.Close()

	return Write(gzw, nbt)
}
