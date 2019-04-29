package formatter

type tokenType int

// we only recognize four types of token now,
// left brace, right brace, end of line and others
const (
	leftBrace tokenType = iota
	rightBrace
	eol
	singleQuoteStr
	doubleQuoteStr
	tripleSingleQuoteStr
	tripleDoubleQuoteStr
	other
)

type token struct {
	tokenType tokenType
	value     string
}
