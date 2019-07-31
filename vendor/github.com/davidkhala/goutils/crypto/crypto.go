package crypto

import "errors"

func assertEmpty(rest []byte, message string) {
	if rest != nil && len(rest) > 0 {
		panic(errors.New(message + ":" + string(rest)))
	}
}