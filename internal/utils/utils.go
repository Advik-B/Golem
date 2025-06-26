package utils

import (
	"fmt"
	"math/big"
)

func JavaHexDigest(b []byte) string {
	// Minecraft uses signed decimal digest
	i := new(big.Int).SetBytes(b)
	if b[0]&0x80 != 0 {
		// negative
		i = i.Sub(i, new(big.Int).Lsh(big.NewInt(1), uint(len(b)*8)))
	}
	return i.Text(16)
}

func FormatUUID(raw string) string {
	return fmt.Sprintf("%s-%s-%s-%s-%s", raw[0:8], raw[8:12], raw[12:16], raw[16:20], raw[20:])
}
