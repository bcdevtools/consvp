package utils

func TruncateStringUntilBufferLessThanXBytesOrFillWithSpaceSuffix(input string, maxBytes int) []byte {
	bzInput := []byte(input)
	for len(bzInput) > maxBytes {
		recoverRune := []rune(string(bzInput))
		recoverRune = recoverRune[:len(recoverRune)-1] // remove last rune
		bzInput = []byte(string(recoverRune))
	}
	bzOutput := make([]byte, maxBytes)
	copy(bzOutput, bzInput)
	for i := 0; i < maxBytes; i++ {
		if bzOutput[i] == 0 {
			bzOutput[i] = ' '
		}
	}
	return bzOutput
}
