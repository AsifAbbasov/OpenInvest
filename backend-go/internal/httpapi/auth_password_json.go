package httpapi

import (
	"encoding/json"
	"fmt"
	"unicode/utf8"
)

type losslessPassword string

func (password *losslessPassword) UnmarshalJSON(raw []byte) error {
	if err := validateJSONPasswordUnicode(raw); err != nil {
		return err
	}
	var decoded string
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return err
	}
	if !utf8.ValidString(decoded) {
		return fmt.Errorf("password must contain valid Unicode")
	}
	*password = losslessPassword(decoded)
	return nil
}

func validateJSONPasswordUnicode(raw []byte) error {
	if len(raw) < 2 || raw[0] != '"' || raw[len(raw)-1] != '"' {
		return fmt.Errorf("password must be a JSON string")
	}
	if !utf8.Valid(raw) {
		return fmt.Errorf("password contains invalid UTF-8")
	}

	for index := 1; index < len(raw)-1; {
		if raw[index] != '\\' {
			_, size := utf8.DecodeRune(raw[index : len(raw)-1])
			index += size
			continue
		}
		if index+1 >= len(raw)-1 {
			return fmt.Errorf("password contains an invalid JSON escape")
		}
		if raw[index+1] != 'u' {
			index += 2
			continue
		}

		codeUnit, ok := parseJSONHex4(raw[index+2:])
		if !ok {
			return fmt.Errorf("password contains an invalid Unicode escape")
		}
		switch {
		case codeUnit >= 0xD800 && codeUnit <= 0xDBFF:
			if index+12 > len(raw)-1 || raw[index+6] != '\\' || raw[index+7] != 'u' {
				return fmt.Errorf("password contains an unpaired high surrogate")
			}
			lowSurrogate, ok := parseJSONHex4(raw[index+8:])
			if !ok || lowSurrogate < 0xDC00 || lowSurrogate > 0xDFFF {
				return fmt.Errorf("password contains an invalid surrogate pair")
			}
			index += 12
		case codeUnit >= 0xDC00 && codeUnit <= 0xDFFF:
			return fmt.Errorf("password contains an unpaired low surrogate")
		default:
			index += 6
		}
	}
	return nil
}

func parseJSONHex4(raw []byte) (uint16, bool) {
	if len(raw) < 4 {
		return 0, false
	}
	var value uint16
	for _, digit := range raw[:4] {
		value <<= 4
		switch {
		case digit >= '0' && digit <= '9':
			value += uint16(digit - '0')
		case digit >= 'a' && digit <= 'f':
			value += uint16(digit-'a') + 10
		case digit >= 'A' && digit <= 'F':
			value += uint16(digit-'A') + 10
		default:
			return 0, false
		}
	}
	return value, true
}
