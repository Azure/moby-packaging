package archive

import (
	"fmt"
	"os"
	"strings"
)

type textKind int

type EText []byte

const EOF = '\x05'

const (
	invalid textKind = iota
	filename
	str
)

func (t text) String() string {
	var x string
	switch t.kind {
	case filename:
		x = "filename"
	case str:
		x = "str"
	default:
		x = "invalid"
	}
	return x + "(" + t.data + ")"
}

type text struct {
	kind textKind
	data string
}

func (t *EText) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}

	text, err := parseText(s)
	if err != nil {
		return err
	}

	var buf []byte
	switch text.kind {
	case filename:
		b, err := os.ReadFile(text.data)
		if err != nil {
			return err
		}
		buf = b
	case str:
		buf = []byte(text.data)
	default:
		return fmt.Errorf("unintelligible string: %s", s)
	}

	*t = buf
	return nil
}

func parseText(s string) (text, error) {
	data := []rune(s)
	var t textKind

	data = append(data, EOF)
	i := 0

	buf := new(strings.Builder)

	state := 0
	for {
		if data[i] == EOF {
			break
		}

		switch state {
		case 0:
			switch data[i] {
			case '\\':
				state = 2
			case '#':
				t = filename
				state = 1
			default:
				t = str
				buf.WriteRune(data[i])
				state = 1
			}
		case 1:
			buf.WriteRune(data[i])
		case 2:
			t = str
			buf.WriteRune(data[i])
			state = 1
		}

		i++
	}

	if t == invalid {
		return text{}, fmt.Errorf("invalid string or filename: %s", s)
	}

	return text{
		kind: t,
		data: buf.String(),
	}, nil
}
