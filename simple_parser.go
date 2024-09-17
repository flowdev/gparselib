package gparselib

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ParseLiteral parses a literal value at the current position of the parser.
// The configuration has to be the literal string we expect.
func ParseLiteral(
	pd *ParseData, ctx interface{},
	pluginSemantics SemanticsOp,
	cfgLiteral string,
) (*ParseData, interface{}) {
	cfgN := len(cfgLiteral)
	pos := pd.Source.pos
	if len(pd.Source.content) >= pos+cfgN &&
		pd.Source.content[pos:pos+cfgN] == cfgLiteral {

		createMatchedResult(pd, cfgN)
	} else {
		createUnmatchedResult(
			pd,
			0,
			"Literal '"+cfgLiteral+"' expected",
			nil)
	}
	return handleSemantics(pluginSemantics, pd, ctx)
}

// NewParseLiteralPlugin creates a plugin sporting a literal parser.
func NewParseLiteralPlugin(pluginSemantics SemanticsOp, cfgLiteral string) SubparserOp {
	return func(pd *ParseData, ctx interface{}) (*ParseData, interface{}) {
		return ParseLiteral(pd, ctx, pluginSemantics, cfgLiteral)
	}
}

// ParseIdent parses an identifier at the current position of the parser.
// If allows Unicode letters for the first character and Unicode letters
// and Unicode numbers for all following characters.
// The configuration has to be the additional characters
// allowed for the first and following characters.
func ParseIdent(
	pd *ParseData, ctx interface{},
	pluginSemantics SemanticsOp,
	cfgFirstChar, cfgFollowingChars string,
) (*ParseData, interface{}, error) {
	var n int
	pos := pd.Source.pos
	substr := pd.Source.content[pos:]

	for {
		r, size := utf8.DecodeRuneInString(substr)
		if r == utf8.RuneError {
			break
		}
		if (unicode.IsLetter(r)) || // letters are always allowed
			(n > 0 && unicode.IsNumber(r)) || // digits are only allowed as following chars
			(n == 0 && strings.ContainsRune(cfgFirstChar, r)) || // configured for first char
			(n > 0 && strings.ContainsRune(cfgFollowingChars, r)) { // configured for following chars

			n += size
			substr = substr[size:]
		} else { // no ident
			break
		}
	}

	if n > 0 {
		createMatchedResult(pd, n)
	} else {
		createUnmatchedResult(pd, 0, "Identifier expected", nil)
	}
	pd, ctx = handleSemantics(pluginSemantics, pd, ctx)
	return pd, ctx, nil
}

// NewParseIdentPlugin creates a plugin sporting an identifier parser.
func NewParseIdentPlugin(pluginSemantics SemanticsOp, cfgFirstChar, cfgFollowingChars string) SubparserOp {
	return func(pd *ParseData, ctx interface{}) (*ParseData, interface{}) {
		pd, ctx, _ = ParseIdent(pd, ctx, pluginSemantics, cfgFirstChar, cfgFollowingChars)
		return pd, ctx
	}
}

// This is needed for: ParseNatural
const allDigits = "0123456789abcdefghijklmnopqrstuvwxyz"

// ParseNatural parses a natural number at the current position of the parser.
// The configuration has to be the radix of accepted numbers (e.g.: 10).
// If the radix is smaller than 2 or larger than 36 an error is returned.
func ParseNatural(
	pd *ParseData, ctx interface{},
	pluginSemantics SemanticsOp,
	cfgRadix int,
) (*ParseData, interface{}, error) {
	if cfgRadix < 2 || cfgRadix > 36 {
		return nil, nil,
			&ParseError{
				where: "",
				myErr: fmt.Sprintf(
					"The radix has to be between 2 and 36, but is: %d",
					cfgRadix,
				),
				baseErr: nil,
			}
	}
	cfgDigits := allDigits[:cfgRadix]

	var n int
	pos := pd.Source.pos
	substr := pd.Source.content[pos:]

	for i, digit := range substr {
		if strings.IndexRune(cfgDigits, unicode.ToLower(digit)) >= 0 {
			n = i + 1
		} else {
			break
		}
	}
	if n > 0 {
		val, err := strconv.ParseUint(substr[:n], cfgRadix, 64)
		if err == nil {
			createMatchedResult(pd, n)
			pd.Result.Value = val
		} else {
			createUnmatchedResult(pd, 0, "Natural number expected", err)
		}
	} else {
		createUnmatchedResult(pd, 0, "Natural number expected", nil)
	}
	pd, ctx = handleSemantics(pluginSemantics, pd, ctx)
	return pd, ctx, nil
}

// NewParseNaturalPlugin creates a plugin sporting a number parser.
func NewParseNaturalPlugin(pluginSemantics SemanticsOp, cfgRadix int) (SubparserOp, error) {
	pd := &ParseData{Source: SourceData{}}
	_, _, err := ParseNatural(pd, nil, nil, cfgRadix)
	if err != nil {
		return nil, err
	}

	return func(pd *ParseData, ctx interface{}) (*ParseData, interface{}) {
		pd, ctx, _ = ParseNatural(pd, ctx, pluginSemantics, cfgRadix)
		return pd, ctx
	}, nil
}

// ParseEOF only matches at the end of the input.
func ParseEOF(
	pd *ParseData, ctx interface{},
	pluginSemantics SemanticsOp,
) (*ParseData, interface{}) {
	pos := pd.Source.pos
	n := len(pd.Source.content)

	if n > pos {
		createUnmatchedResult(pd, 0,
			fmt.Sprintf(
				"Expecting end of input but still got %d bytes",
				n-pos,
			),
			nil,
		)
	} else {
		createMatchedResult(pd, 0)
	}
	return handleSemantics(pluginSemantics, pd, ctx)
}

// NewParseEOFPlugin creates a plugin sporting an EOF parser.
func NewParseEOFPlugin(pluginSemantics SemanticsOp) SubparserOp {
	return func(pd *ParseData, ctx interface{}) (*ParseData, interface{}) {
		return ParseEOF(pd, ctx, pluginSemantics)
	}
}

// ParseSpace parses one or more space characters.
// Space is defined by unicode.IsSpace().
// It can be configured wether EOL ('\n') is to be interpreted as space or not.
func ParseSpace(
	pd *ParseData, ctx interface{},
	pluginSemantics SemanticsOp,
	cfgEOLOK bool,
) (*ParseData, interface{}) {
	var n int
	pos := pd.Source.pos
	substr := pd.Source.content[pos:]

	for {
		r, size := utf8.DecodeRuneInString(substr)
		if r == utf8.RuneError {
			break
		}
		if unicode.IsSpace(r) && (cfgEOLOK || r != '\n') {
			n += size
			substr = substr[size:]
		} else {
			break
		}
	}
	if n > 0 {
		createMatchedResult(pd, n)
	} else {
		createUnmatchedResult(pd, 0, "Expecting white space", nil)
	}
	return handleSemantics(pluginSemantics, pd, ctx)
}

// NewParseSpacePlugin creates a plugin sporting a space parser.
func NewParseSpacePlugin(pluginSemantics SemanticsOp, cfgEOLOK bool) SubparserOp {
	return func(pd *ParseData, ctx interface{}) (*ParseData, interface{}) {
		return ParseSpace(pd, ctx, pluginSemantics, cfgEOLOK)
	}
}

// RegexpParser parses text according to a predefined regular expression.
// The regular expression (e.g.: `^[a-z]+`) has to be configured.
// If the regular expression doesn't start with a `^` it will be added
// automatically.
// If the regular expression can't be compiled an error is returned.
type RegexpParser regexp.Regexp

// NewRegexpParser creates a new parser for the given regular expression.
// If the regular expression is invalid an error is returned.
func NewRegexpParser(cfgRegexp string) (*RegexpParser, error) {
	if cfgRegexp[0] != '^' {
		cfgRegexp = "^" + cfgRegexp
	}
	re, err := regexp.Compile(cfgRegexp)
	return (*RegexpParser)(re), err
}

// ParseRegexp is the input port of the RegexpParser operation.
func (pr *RegexpParser) ParseRegexp(
	pd *ParseData, ctx interface{},
	pluginSemantics SemanticsOp,
) (*ParseData, interface{}) {
	re := (*regexp.Regexp)(pr)
	pos := pd.Source.pos
	substr := pd.Source.content[pos:]
	match := re.FindStringIndex(substr)

	if match != nil {
		createMatchedResult(pd, match[1])
		pd.Result.Value = pd.Result.Text
	} else {
		createUnmatchedResult(
			pd,
			0,
			"Expecting match for regexp `"+re.String()[1:]+"`",
			nil,
		)
	}
	return handleSemantics(pluginSemantics, pd, ctx)
}

// NewParseRegexpPlugin creates a plugin sporting a regular expression parser.
func NewParseRegexpPlugin(
	pluginSemantics SemanticsOp,
	cfgRegexp string,
) (SubparserOp, error) {
	pr, err := NewRegexpParser(cfgRegexp)
	if err != nil {
		return nil, err
	}

	return func(pd *ParseData, ctx interface{}) (*ParseData, interface{}) {
		return pr.ParseRegexp(pd, ctx, pluginSemantics)
	}, nil
}

// ParseLineComment parses a comment until the end of the line.
// The string that starts the comment (e.g.: `//`) has to be configured.
// If the start of the comment is empty an error is returned.
func ParseLineComment(
	pd *ParseData, ctx interface{},
	pluginSemantics SemanticsOp,
	cfgStart string,
) (*ParseData, interface{}, error) {
	if cfgStart == "" {
		return nil, nil,
			errors.New(
				"expected start of line comment as config, got empty string",
			)
	}

	pos := pd.Source.pos
	l := len(cfgStart)
	n := min(pos+l, len(pd.Source.content))
	substr := pd.Source.content[pos:n]

	if substr == cfgStart {
		i := strings.IndexRune(pd.Source.content[n:], '\n')
		if i >= 0 {
			l += i
		} else {
			l = len(pd.Source.content) - pos
		}
		createMatchedResult(pd, l)
		pd.Result.Value = ""
	} else {
		createUnmatchedResult(pd, 0, "Expecting line comment", nil)
	}
	pd, ctx = handleSemantics(pluginSemantics, pd, ctx)
	return pd, ctx, nil
}

// NewParseLineCommentPlugin creates a plugin sporting a number parser.
func NewParseLineCommentPlugin(
	pluginSemantics SemanticsOp,
	cfgStart string,
) (SubparserOp, error) {
	pd := &ParseData{Source: SourceData{}}
	_, _, err := ParseLineComment(pd, nil, nil, cfgStart)
	if err != nil {
		return nil, err
	}

	return func(pd *ParseData, ctx interface{}) (*ParseData, interface{}) {
		pd, ctx, _ = ParseLineComment(pd, ctx, pluginSemantics, cfgStart)
		return pd, ctx
	}, nil
}

// ParseBlockComment parses a comment until the end of the line.
// The strings that start and end the comment (e.g.: `/*`, `*/`)
// have to be configured.
// A comment start or end inside a string literal (', " and `) is ignored.
// If the start or end of the comment is empty an error is returned.
func ParseBlockComment(
	pd *ParseData, ctx interface{},
	pluginSemantics SemanticsOp,
	cfgStart, cfgEnd string,
) (*ParseData, interface{}, error) {
	if cfgStart == "" {
		return nil, nil,
			errors.New(
				"expected start of block comment as config, got empty string",
			)
	}
	if cfgEnd == "" {
		return nil, nil,
			errors.New(
				"expected end of block comment as config, got empty string",
			)
	}
	lBeg := len(cfgStart)
	lEnd := len(cfgEnd)

	pos := pd.Source.pos
	n := min(pos+lBeg, len(pd.Source.content))
	substr := pd.Source.content[pos:n]

	if substr == cfgStart {
		afterBackslash := false
		stringType := ' '
		found := false
		endRune, _ := utf8.DecodeRuneInString(cfgEnd)
		reststr := pd.Source.content[n:]

	RuneLoop:
		for i, r := range reststr {
			switch {
			case afterBackslash:
				afterBackslash = false
			case stringType == '\'' || stringType == '"':
				switch r {
				case '\\':
					afterBackslash = true
				case stringType:
					stringType = ' '
				}
			case stringType == '`':
				if r == '`' {
					stringType = ' '
				}
			default:
				switch r {
				case '\'', '"', '`':
					stringType = r
				case endRune:
					if len(reststr) >= i+lEnd &&
						reststr[i:i+lEnd] == cfgEnd {

						found = true
						pos = i + lEnd
						break RuneLoop
					}
				}
			}
		}
		if found {
			createMatchedResult(pd, lBeg+pos)
			pd.Result.Value = ""
		} else {
			createUnmatchedResult(
				pd,
				lBeg,
				fmt.Sprintf("Block comment isn't closed with '%s'", cfgEnd),
				nil,
			)
			pd.Source.pos += lBeg
		}
	} else {
		createUnmatchedResult(
			pd,
			0,
			fmt.Sprintf(
				"Expecting block comment starting with '%s', got '%s'",
				cfgStart,
				substr),
			nil,
		)
	}
	pd, ctx = handleSemantics(pluginSemantics, pd, ctx)
	return pd, ctx, nil
}

// NewParseBlockCommentPlugin creates a plugin sporting a number parser.
func NewParseBlockCommentPlugin(
	pluginSemantics SemanticsOp,
	cfgStart, cfgEnd string,
) (SubparserOp, error) {
	pd := &ParseData{Source: SourceData{}}
	_, _, err := ParseBlockComment(pd, nil, nil, cfgStart, cfgEnd)
	if err != nil {
		return nil, err
	}

	return func(pd *ParseData, ctx interface{}) (*ParseData, interface{}) {
		pd, ctx, _ = ParseBlockComment(pd, ctx, pluginSemantics, cfgStart, cfgEnd)
		return pd, ctx
	}, nil
}
