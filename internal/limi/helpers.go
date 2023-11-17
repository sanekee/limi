package limi

// splitEscape split string by deliminitor allowing escape character
func SplitEscape(str string, delim byte) []string {
	var escape int
	var splitted []string
	var curStr []byte
	for i := 0; i < len(str); i++ {
		c := byte(str[i])
		if c == '\\' && i-escape > 1 {
			escape = i
			continue
		}
		if c == delim && i-escape > 1 {
			splitted = append(splitted, string(curStr))
			curStr = []byte{}
			continue
		}

		curStr = append(curStr, c)
	}
	if len(curStr) > 0 {
		splitted = append(splitted, string(curStr))
	}
	return splitted
}
