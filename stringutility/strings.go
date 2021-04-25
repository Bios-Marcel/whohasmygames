package stringutility

import "strings"

// StringJoiner is similar to strings.Builder, but it puts a separator rune
// between items. The separator won't be added before the first item or after
// the last item.
type StringJoiner struct {
	*strings.Builder
	joiner rune
}

// NewStringJoiner creates a ready to use StringJoiner. The passed joiner rune
// will be aplied after each WriteString call, but not after WriteRune or
// WriteByte.
func NewStringJoiner(joiner rune) *StringJoiner {
	return &StringJoiner{
		Builder: &strings.Builder{},
		joiner:  joiner,
	}
}

// WriteString adds the string to the underlying StringBuilder and optionally
// adds the joiner rune if necessary.
func (sj *StringJoiner) WriteString(s string) {
	if sj.Len() != 0 {
		sj.WriteRune(sj.joiner)
	}

	sj.Builder.WriteString(s)
}
