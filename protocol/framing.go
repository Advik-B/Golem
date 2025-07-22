package protocol

import (
	"bytes"
	"errors"
	"io"
)

// Framer is a writer that prepends the length of the data as a VarInt.
type Framer struct {
	w io.Writer
}

// NewFramer creates a new Framer.
func NewFramer(w io.Writer) *Framer {
	return &Framer{w: w}
}

// WritePacket takes a complete packet payload, frames it, and writes it to the underlying writer.
func (f *Framer) WritePacket(payload []byte) error {
	var buf bytes.Buffer
	if err := WriteVarInt(&buf, int32(len(payload))); err != nil {
		return err
	}
	if _, err := buf.Write(payload); err != nil {
		return err
	}
	_, err := f.w.Write(buf.Bytes())
	return err
}

// Splitter is a reader that reads length-prefixed packets from an underlying reader.
type Splitter struct {
	r      io.Reader
	buffer bytes.Buffer
}

// NewSplitter creates a new Splitter.
func NewSplitter(r io.Reader) *Splitter {
	return &Splitter{r: r}
}

// NextPacket reads the next full packet payload from the stream.
// It handles partial reads by buffering data until a complete packet is available.
func (s *Splitter) NextPacket() ([]byte, error) {
	for {
		// Create a reader from the current buffer to attempt a read.
		bufReader := bytes.NewReader(s.buffer.Bytes())

		// Attempt to read the packet length.
		packetLen, err := ReadVarInt(bufReader)
		if err != nil {
			// If the error is EOF, it means we don't have enough data for the VarInt.
			// We need to read more from the underlying connection.
			if err == io.EOF || errors.Is(err, io.ErrUnexpectedEOF) {
				if err := s.fillBuffer(); err != nil {
					return nil, err
				}
				continue // Retry reading the packet length.
			}
			return nil, err // A different error occurred.
		}

		// Calculate how many bytes the VarInt itself took up.
		varIntLen := s.buffer.Len() - bufReader.Len()

		// Check if we have the full packet in the buffer.
		if int32(s.buffer.Len()-varIntLen) >= packetLen {
			// We have a full packet, extract it.
			start := varIntLen
			end := varIntLen + int(packetLen)
			packetData := s.buffer.Bytes()[start:end]

			// Reset the buffer to the remaining data.
			s.buffer.Next(end)

			return packetData, nil
		}

		// Not enough data for the full packet, read more from the network.
		if err := s.fillBuffer(); err != nil {
			// If we get an EOF here, it means the connection closed mid-packet.
			if err == io.EOF {
				return nil, io.ErrUnexpectedEOF
			}
			return nil, err
		}
	}
}

// fillBuffer reads data from the underlying reader and appends it to the internal buffer.
func (s *Splitter) fillBuffer() error {
	// Use a temporary buffer to avoid resizing the main buffer on every small read.
	tmp := make([]byte, 4096)
	n, err := s.r.Read(tmp)
	if n > 0 {
		s.buffer.Write(tmp[:n])
	}

	// An EOF is only an error if no bytes were read. If we read some bytes and
	// then get an EOF, that's a valid state. The NextPacket loop will handle it.
	if err == io.EOF && n > 0 {
		return nil
	}

	return err
}
