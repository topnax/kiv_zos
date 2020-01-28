package utils

// returns true when the byte has a set bit at the given position
func HasBit(n byte, pos int8) bool {
	return n&(1<<pos) > 0
}

// sets a bit to 1 at the given position
func SetBit(n byte, pos int8) byte {
	return n | (1 << pos)
}

// sets a bit to 0 at the given position
func ClearBit(n byte, pos int8) byte {
	return n & ^(1 << pos)
}
