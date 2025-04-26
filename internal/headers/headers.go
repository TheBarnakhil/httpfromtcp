package headers

import (
	"bytes"
	"errors"
	"strings"
	"unicode"
)

type Headers map[string]string

const crlf = "\r\n"

var specialChars = []byte{'!', '#', '$', '%', '&', '*', '+', '-', '.', '^', '_', '`', '|', '~', '\''}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return 0, false, nil
	} else if idx == 0 {
		return 2, true, nil
	}
	data = data[:idx]

	keyVal := bytes.SplitN(data, []byte(":"), 2)
	if len(keyVal) != 2 {
		return 0, false, errors.New("error: Found no ':' in the header")
	}

	key, value := keyVal[0], keyVal[1]
	if bytes.HasSuffix(key, []byte(" ")) {
		return 0, false, errors.New("error: Found whitespace between colon and key, invalid format")
	}

	key = bytes.TrimSpace(key)
	for _, char := range key {
		if !(unicode.IsUpper(rune(char)) || unicode.IsDigit(rune(char)) || unicode.IsLower(rune(char)) || bytes.ContainsRune(specialChars, rune(char))) {
			return 0, false, errors.New("error: invalid header key")
		}
	}

	value = bytes.TrimSpace(value)

	mapKey := string(bytes.ToLower(key))

	if val, exists := h[mapKey]; exists {
		h[mapKey] = val + ", " + string(value)
	} else {
		h[mapKey] = string(value)
	}

	return idx + 2, false, nil
}

func NewHeaders() Headers {
	return Headers{}
}

func (h Headers) Get(key string) (string, bool) {
	v, ok := h[strings.ToLower(key)]
	return v, ok
}
