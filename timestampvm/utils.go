package timestampvm

// BytesToData converts a byte slice to an array. If the byte slice input is
// larger than [DataLen], it will be truncated.
func BytesToData(input []byte) [DataLen]byte {
	data := [DataLen]byte{}
	lim := len(input)
	if lim > DataLen {
		lim = DataLen
	}
	copy(data[:], input[:lim])
	return data
}
