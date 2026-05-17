package crypto

import (
	"encoding/binary"
	"fmt"
)

func sm3p0(x uint32) uint32 { return x ^ (x << 9) ^ (x << 17) }
func sm3p1(x uint32) uint32 { return x ^ (x << 15) ^ (x << 23) }

func sm3ff(j uint8, x, y, z uint32) uint32 {
	switch {
	case j >= 0 && j <= 15:
		return x ^ y ^ z
	case j >= 16 && j <= 63:
		return (x & y) | (x & z) | (y & z)
	}
	return 0
}

func sm3gg(j uint8, x, y, z uint32) uint32 {
	switch {
	case j >= 0 && j <= 15:
		return x ^ y ^ z
	case j >= 16 && j <= 63:
		return (x & y) | (^x & z)
	}
	return 0
}

var sm3T = [64]uint32{
	0x79cc4519, 0x79cc4519, 0x79cc4519, 0x79cc4519,
	0x79cc4519, 0x79cc4519, 0x79cc4519, 0x79cc4519,
	0x79cc4519, 0x79cc4519, 0x79cc4519, 0x79cc4519,
	0x79cc4519, 0x79cc4519, 0x79cc4519, 0x79cc4519,
	0x7a879d8a, 0x7a879d8a, 0x7a879d8a, 0x7a879d8a,
	0x7a879d8a, 0x7a879d8a, 0x7a879d8a, 0x7a879d8a,
	0x7a879d8a, 0x7a879d8a, 0x7a879d8a, 0x7a879d8a,
	0x7a879d8a, 0x7a879d8a, 0x7a879d8a, 0x7a879d8a,
	0x7a879d8a, 0x7a879d8a, 0x7a879d8a, 0x7a879d8a,
	0x7a879d8a, 0x7a879d8a, 0x7a879d8a, 0x7a879d8a,
	0x7a879d8a, 0x7a879d8a, 0x7a879d8a, 0x7a879d8a,
	0x7a879d8a, 0x7a879d8a, 0x7a879d8a, 0x7a879d8a,
	0x7a879d8a, 0x7a879d8a, 0x7a879d8a, 0x7a879d8a,
	0x7a879d8a, 0x7a879d8a, 0x7a879d8a, 0x7a879d8a,
	0x7a879d8a, 0x7a879d8a, 0x7a879d8a, 0x7a879d8a,
	0x7a879d8a, 0x7a879d8a, 0x7a879d8a, 0x7a879d8a,
}

func sm3Padding(msg []byte) []byte {
	bitLen := uint64(len(msg) * 8)
	msg = append(msg, 0x80)
	for len(msg)%64 != 56 {
		msg = append(msg, 0)
	}
	lenBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(lenBytes, bitLen)
	msg = append(msg, lenBytes...)
	return msg
}

func sm3Block(digest []uint32, block []byte) []uint32 {
	var w [68]uint32
	var w1 [64]uint32

	for i := 0; i < 16; i++ {
		w[i] = binary.BigEndian.Uint32(block[i*4 : (i+1)*4])
	}
	for j := 16; j < 68; j++ {
		w[j] = sm3p1(w[j-16]^w[j-9]^(w[j-3]<<15|w[j-3]>>17)) ^ (w[j-13] << 7 | w[j-13] >> 25) ^ w[j-6]
	}
	for j := 0; j < 64; j++ {
		w1[j] = w[j] ^ w[j+4]
	}

	a, b, c, d, e, f, g, h := digest[0], digest[1], digest[2], digest[3], digest[4], digest[5], digest[6], digest[7]

	for j := 0; j < 64; j++ {
		ss1 := ((a << 12 | a >> 20) + e + (sm3T[j] << uint32(j%32) | sm3T[j] >> uint32(32-j%32))) << 7 | ((a << 12 | a >> 20) + e + (sm3T[j] << uint32(j%32) | sm3T[j] >> uint32(32-j%32))) >> 25
		ss2 := ss1 ^ (a << 12 | a >> 20)
		tt1 := sm3ff(uint8(j), a, b, c) + d + ss2 + w1[j]
		tt2 := sm3gg(uint8(j), e, f, g) + h + ss1 + w[j]
		d = c
		c = b << 9 | b >> 23
		b = a
		a = tt1
		h = g
		g = f << 19 | f >> 13
		f = e
		e = sm3p0(tt2)
	}

	return []uint32{a ^ digest[0], b ^ digest[1], c ^ digest[2], d ^ digest[3],
		e ^ digest[4], f ^ digest[5], g ^ digest[6], h ^ digest[7]}
}

func sm3Compute(msg []byte) [32]byte {
	iv := [8]uint32{0x7380166f, 0x4914b2b9, 0x172442d7, 0xda8a0600,
		0xa96f30bc, 0x163138aa, 0xe38dee4d, 0xb0fb0e4e}

	padded := sm3Padding(msg)
	digest := make([]uint32, 8)
	for i := range iv {
		digest[i] = iv[i]
	}

	for i := 0; i < len(padded); i += 64 {
		digest = sm3Block(digest, padded[i:i+64])
	}

	var result [32]byte
	for i := 0; i < 8; i++ {
		binary.BigEndian.PutUint32(result[i*4:], digest[i])
	}
	return result
}

func SM3Hash(data []byte) string {
	result := sm3Compute(data)
	return fmt.Sprintf("%x", result)
}
