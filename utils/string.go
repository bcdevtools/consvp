package utils

func TruncateStringUntilBufferLessThanXBytesOrFillWithSpaceSuffix(s string, maxBytes int) []byte {
	bz := []byte(s)
	for len(bz) > maxBytes {
		recoverRune := []rune(string(bz))
		recoverRune = recoverRune[:len(recoverRune)-1] // remove last rune
		bz = []byte(string(recoverRune))
	}
	resBz := make([]byte, maxBytes)
	copy(resBz, bz)
	for i := 0; i < maxBytes; i++ {
		if resBz[i] == 0 {
			resBz[i] = ' '
		}
	}
	return resBz
}
