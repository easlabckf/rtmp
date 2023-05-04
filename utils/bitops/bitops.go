package bitops

func U8(b []byte) (i uint8) {
	return b[0]
}

func U32BE(b []byte) (i uint32) {
	i = uint32(b[0])
	i <<= 8
	i |= uint32(b[1])
	i <<= 8
	i |= uint32(b[2])
	i <<= 8
	i |= uint32(b[3])
	return
}
