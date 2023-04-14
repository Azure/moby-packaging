package archive

import (
	"os"
	"strings"
)

type Text string

func (t *Text) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}

	_, escaped := strings.CutPrefix(s, `\`)

	filenameOrText, ok := strings.CutPrefix(s, "#")
	if ok && !escaped {

		b, err := os.ReadFile(filenameOrText)
		if err != nil {
			return err
		}

		*t = Text(b)
		return nil
	}

	*t = Text(filenameOrText)
	return nil
}
