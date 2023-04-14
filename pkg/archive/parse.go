package archive

import (
	"os"
	"strings"
)

type Text string
type TextList []Text

func (t Text) String() string {
	return string(t)
}

func (tl TextList) Strings() []string {
	ret := []string{}

	for _, s := range tl {
		ret = append(ret, s.String())
	}

	return ret
}

func (t *Text) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}

	var escaped bool
	s, escaped = strings.CutPrefix(s, `\`)

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
