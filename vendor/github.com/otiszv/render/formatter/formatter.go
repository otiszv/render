package formatter

import (
	"regexp"
	"strings"
)

const defaultIndentSpaces = 4

func Format(jenkinsfile string) string {
	indentLevel := 0
	scanner := scanner{
		chars: []rune(jenkinsfile),
		pos:   0,
	}

	var result strings.Builder

	tokens := scanner.readAllToTokens()
	for i := 0; i < len(tokens); i++ {
		currentToken := tokens[i]
		var nextToken *token
		if i == len(tokens)-1 {
			nextToken = nil
		} else {
			nextToken = &tokens[i+1]
		}

		switch currentToken.tokenType {
		case leftBrace:
			result.WriteString(currentToken.value)
			indentLevel++

			if nextToken == nil {
				result.WriteString("\n")
			} else if nextToken.tokenType == eol {
				break
			} else {
				result.WriteString("\n")
				result.WriteString(indent(indentLevel))
			}
		case rightBrace:
			result.WriteString("\n")
			indentLevel--
			result.WriteString(indent(indentLevel))
			result.WriteString(currentToken.value)

			if nextToken == nil {
				result.WriteString("\n")
			} else if nextToken.tokenType == other {
				result.WriteString("\n")
				result.WriteString(indent(indentLevel))
			}
		case eol:
			if nextToken == nil || nextToken.tokenType == eol || nextToken.tokenType == rightBrace || (nextToken.tokenType == other && strings.TrimSpace(nextToken.value) == "") {
				break
			}

			result.WriteString(currentToken.value)
			result.WriteString(indent(indentLevel))
		case other, singleQuoteStr, doubleQuoteStr, tripleSingleQuoteStr, tripleDoubleQuoteStr:
			if currentToken.tokenType == other {
				result.WriteString(strings.TrimLeft(currentToken.value, " \t"))
			} else {
				result.WriteString(currentToken.value)
			}
			if nextToken != nil && nextToken.tokenType == leftBrace && !strings.HasSuffix(currentToken.value, " ") {
				result.WriteString(" ")
			}
		}
	}

	return result.String()
}

func removeSpaceLine(jenkinsfile string) string {
	re := regexp.MustCompile("(?m)^[\\s]*$[\r\n]*")

	return strings.Trim(re.ReplaceAllString(jenkinsfile, ""), "\r\n")
}

func indent(indentLevel int) string {
	return strings.Repeat(" ", indentLevel*defaultIndentSpaces)
}
