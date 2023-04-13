package archive

import (
	"fmt"
	"os"
	"strings"
)

type kindOfText int

const (
	EOF = '\x05'

	invalid kindOfText = iota
	filename
	str
)

type textOrFile struct {
	kind kindOfText
	data string
}

type Text string

func (t *Text) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}

	text, err := parseText(s)
	if err != nil {
		return err
	}

	var buf string
	switch text.kind {
	case filename:
		b, err := os.ReadFile(text.data)
		if err != nil {
			return err
		}

		buf = string(b)
	case str:
		buf = text.data
	default:
		return fmt.Errorf("unintelligible string: %s", s)
	}

	*t = Text(buf)
	return nil
}

func parseText(s string) (textOrFile, error) {
	data := []rune(s)
	var t kindOfText

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
		return textOrFile{}, fmt.Errorf("invalid string or filename: %s", s)
	}

	return textOrFile{
		kind: t,
		data: buf.String(),
	}, nil
}
