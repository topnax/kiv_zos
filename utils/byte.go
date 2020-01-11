package utils

func HasBit(n byte, pos int8) bool {
	return n&(1<<pos) > 0
}

func SetBit(n byte, pos int8) byte {
	return n | (1 << pos)
}

func ClearBit(n byte, pos int8) byte {
	return n & ^(1 << pos)
}
