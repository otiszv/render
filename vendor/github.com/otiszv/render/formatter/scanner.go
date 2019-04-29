package formatter

import "errors"

type scanner struct {
	pos   int
	chars []rune
}

func (s *scanner) readAllToTokens() []token {
	tokens := make([]token, 0)

	for {
		t, err := s.readToken()
		if err != nil {
			break
		}

		// merge string type and other type to other type if last token is other type
		if len(tokens) > 0 && tokens[len(tokens)-1].tokenType == other &&
			(t.tokenType == singleQuoteStr || t.tokenType == doubleQuoteStr ||
				t.tokenType == tripleDoubleQuoteStr || t.tokenType == tripleSingleQuoteStr || t.tokenType == other) {
			lastToken := tokens[len(tokens)-1]
			tokens[len(tokens)-1] = token{
				tokenType: other,
				value:     lastToken.value + t.value,
			}
		} else {
			tokens = append(tokens, t)
		}
	}

	return tokens
}

func (s *scanner) readToken() (token, error) {
	startPos := s.pos

	for ; s.pos < len(s.chars); s.pos++ {
		currentChar := s.chars[s.pos]

		switch currentChar {
		case '{', '}':
			if startPos != s.pos {
				return token{
					tokenType: other,
					value:     string(s.chars[startPos:s.pos]),
				}, nil
			}

			var tokenType tokenType
			if currentChar == '{' {
				tokenType = leftBrace
			} else {
				tokenType = rightBrace
			}

			s.pos++
			return token{
				tokenType: tokenType,
				value:     string(currentChar),
			}, nil
		case '\'':
			if startPos != s.pos {
				return token{
					tokenType: other,
					value:     string(s.chars[startPos:s.pos]),
				}, nil
			}

			if s.isEscaped(s.pos) {
				continue
			} else if s.isTripleQuote(s.pos, '\'') {
				// if is triple quote, the content between this and next triple quote will be view as triple quote string
				s.pos = s.pos + 3
				for s.pos < len(s.chars) {
					if s.isTripleQuote(s.pos, '\'') {
						s.pos = s.pos + 3
						return token{
							tokenType: tripleSingleQuoteStr,
							value:     string(s.chars[startPos:s.pos]),
						}, nil
					}
					s.pos++
				}

			} else {
				// if is single quote, the content between this and next single quote will be view as single quote string
				for s.pos < len(s.chars) {
					s.pos++

					if s.chars[s.pos] == '\'' && !s.isEscaped(s.pos) {
						s.pos++
						return token{
							tokenType: singleQuoteStr,
							value:     string(s.chars[startPos:s.pos]),
						}, nil
					}
				}
			}
		case '"':
			if startPos != s.pos {
				return token{
					tokenType: other,
					value:     string(s.chars[startPos:s.pos]),
				}, nil
			}

			if s.isEscaped(s.pos) {
				continue
			} else if s.isTripleQuote(s.pos, '"') {
				s.pos = s.pos + 3
				for s.pos < len(s.chars) {
					if s.isTripleQuote(s.pos, '"') {
						s.pos = s.pos + 3
						return token{
							tokenType: tripleDoubleQuoteStr,
							value:     string(s.chars[startPos:s.pos]),
						}, nil
					}
					s.pos++
				}
			} else {
				// if is double quote, the content between this and next single quote will be view as single quote string
				for s.pos < len(s.chars) {
					s.pos++
					if s.pos == len(s.chars) {
						return token{
							tokenType: other,
							value:     string(s.chars[startPos:s.pos]),
						}, nil
					}

					if s.chars[s.pos] == '"' && !s.isEscaped(s.pos) {
						s.pos++
						return token{
							tokenType: doubleQuoteStr,
							value:     string(s.chars[startPos:s.pos]),
						}, nil
					}
				}
			}
		case '\n':
			if startPos != s.pos {
				return token{
					tokenType: other,
					value:     string(s.chars[startPos:s.pos]),
				}, nil
			}

			s.pos++
			return token{
				tokenType: eol,
				value:     "\n",
			}, nil
		}
	}

	if startPos < len(s.chars) {
		return token{
			tokenType: other,
			value:     string(s.chars[startPos:]),
		}, nil
	} else {
		return token{}, errors.New("read to end of file")
	}
}

func (s *scanner) isEscaped(currentPos int) bool {
	escaped := 0

	for pos := currentPos - 1; pos >= 0; pos-- {
		if s.chars[pos] == '\\' {
			escaped++
		} else {
			break
		}
	}

	return escaped%2 != 0
}

func (s *scanner) isTripleQuote(currentPos int, quote rune) bool {
	return currentPos < len(s.chars)-2 &&
		(currentPos == 0 || !s.isEscaped(currentPos)) &&
		s.chars[currentPos] == quote && s.chars[currentPos+1] == quote && s.chars[currentPos+2] == quote
}
