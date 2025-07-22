package protocol

import (
    "bytes"
    "compress/zlib"
    "crypto/aes"
    "crypto/cipher"
    "fmt"
    "io"
)

// --- Compression ---

// Decompressor reads compressed packet data.
type Decompressor struct {
    r         io.Reader
    threshold int32
    zlib      io.ReadCloser
}

func NewDecompressor(r io.Reader, threshold int32) *Decompressor {
    return &Decompressor{r: r, threshold: threshold}
}

// Read reads and decompresses data from the underlying reader.
func (d *Decompressor) Read(p []byte) (n int, err error) {
    dataLength, err := ReadVarInt(d.r)
    if err != nil {
        return 0, err
    }

    if dataLength == 0 {
        // Data is not compressed, read directly.
        return d.r.Read(p)
    }

    if dataLength < d.threshold {
        return 0, fmt.Errorf("server sent compressed packet smaller than threshold: len=%d, threshold=%d", dataLength, d.threshold)
    }

    if d.zlib == nil {
        d.zlib, err = zlib.NewReader(d.r)
        if err != nil {
            return 0, err
        }
    } else {
        // Reset the reader to the underlying stream for the next compressed payload.
        err = d.zlib.(zlib.Resetter).Reset(d.r, nil)
        if err != nil {
            return 0, err
        }
    }
    defer d.zlib.Close()
    return io.ReadFull(d.zlib, p)
}

// Compressor writes compressed packet data.
type Compressor struct {
    w         io.Writer
    threshold int32
    zlib      *zlib.Writer
    buf       bytes.Buffer
}

func NewCompressor(w io.Writer, threshold int32) *Compressor {
    return &Compressor{w: w, threshold: threshold, zlib: zlib.NewWriter(nil)}
}

// Write compresses and writes data to the underlying writer.
func (c *Compressor) Write(p []byte) (n int, err error) {
    if int32(len(p)) < c.threshold {
        // Data is smaller than threshold, write uncompressed.
        if err := WriteVarInt(c.w, 0); err != nil {
            return 0, err
        }
        return c.w.Write(p)
    }

    c.buf.Reset()
    c.zlib.Reset(&c.buf)

    if _, err := c.zlib.Write(p); err != nil {
        return 0, err
    }
    if err := c.zlib.Close(); err != nil {
        return 0, err
    }

    compressedData := c.buf.Bytes()
    if err := WriteVarInt(c.w, int32(len(p))); err != nil {
        return 0, err
    }
    if _, err := c.w.Write(compressedData); err != nil {
        return 0, err
    }
    return len(p), nil
}

// --- Encryption ---

// newCFB8 creates a new Cipher Feedback stream with a segment size of 8 bits.
// Go's standard library CFB is for a segment size equal to the block size (128 bits for AES).
// We must implement CFB8 manually for Minecraft compatibility.
func newCFB8(block cipher.Block, iv []byte, decrypt bool) cipher.Stream {
    return &cfb8Stream{
        block:     block,
        blockSize: block.BlockSize(),
        iv:        append([]byte{}, iv...),
        out:       make([]byte, block.BlockSize()),
        decrypt:   decrypt,
    }
}

type cfb8Stream struct {
    block     cipher.Block
    blockSize int
    iv        []byte
    out       []byte
    decrypt   bool
}

func (x *cfb8Stream) XORKeyStream(dst, src []byte) {
    for i := 0; i < len(src); i++ {
        x.block.Encrypt(x.out, x.iv)

        if x.decrypt {
            copy(x.iv, append(x.iv[1:], src[i]))
        }

        dst[i] = src[i] ^ x.out[0]

        if !x.decrypt {
            copy(x.iv, append(x.iv[1:], dst[i]))
        }
    }
}

// Decipher is an io.Reader that decrypts data from an underlying reader.
type Decipher struct {
    r io.Reader
    s cipher.Stream
}

func NewDecipher(r io.Reader, sharedSecret []byte) (*Decipher, error) {
    block, err := aes.NewCipher(sharedSecret)
    if err != nil {
        return nil, err
    }
    return &Decipher{
        r: r,
        s: newCFB8(block, sharedSecret, true),
    }, nil
}

func (d *Decipher) Read(p []byte) (n int, err error) {
    n, err = d.r.Read(p)
    if n > 0 {
        d.s.XORKeyStream(p[:n], p[:n])
    }
    return
}

// Cipher is an io.Writer that encrypts data to an underlying writer.
type Cipher struct {
    w io.Writer
    s cipher.Stream
}

func NewCipher(w io.Writer, sharedSecret []byte) (*Cipher, error) {
    block, err := aes.NewCipher(sharedSecret)
    if err != nil {
        return nil, err
    }
    return &Cipher{
        w: w,
        s: newCFB8(block, sharedSecret, false),
    }, nil
}

func (c *Cipher) Write(p []byte) (n int, err error) {
    encrypted := make([]byte, len(p))
    c.s.XORKeyStream(encrypted, p)
    return c.w.Write(encrypted)
}
