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
	portOut func(interface{}),
	fillSemantics SemanticsOp,
	getParseData GetParseData,
	setParseData SetParseData,
	cfgLiteral string,
) (portIn func(interface{})) {
	cfgN := len(cfgLiteral)

	portSemOut := makeSemanticsPort(fillSemantics, portOut)
	portIn = func(data interface{}) {
		pd := getParseData(data)
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
		handleSemantics(portOut, portSemOut, setParseData, data, pd)
	}
	return
}

// This is needed for: ParseNatural
const allDigits = "0123456789abcdefghijklmnopqrstuvwxyz"

// ParseNatural parses a natural number at the current position of the parser.
// The configuration has to be the radix of accepted numbers (e.g.: 10).
// If the radix is smaller than 2 or larger than 36 an error is returned.
func ParseNatural(
	portOut func(interface{}),
	fillSemantics SemanticsOp,
	getParseData GetParseData,
	setParseData SetParseData,
	cfgRadix int,
) (portIn func(interface{}), err error) {
	if cfgRadix < 2 || cfgRadix > 36 {
		return nil,
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

	portSemOut := makeSemanticsPort(fillSemantics, portOut)
	portIn = func(data interface{}) {
		var n int
		pd := getParseData(data)
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
			val, err := strconv.ParseUint(substr[:n], len(cfgDigits), 64)
			if err == nil {
				createMatchedResult(pd, n)
				pd.Result.Value = val
			} else {
				createUnmatchedResult(pd, 0, "Natural number expected", err)
			}
		} else {
			createUnmatchedResult(pd, 0, "Natural number expected", nil)
		}
		handleSemantics(portOut, portSemOut, setParseData, data, pd)
	}
	return
}

// ParseEOF only matches at the end of the input.
func ParseEOF(
	portOut func(interface{}),
	fillSemantics SemanticsOp,
	getParseData GetParseData,
	setParseData SetParseData,
) (portIn func(interface{})) {
	portSemOut := makeSemanticsPort(fillSemantics, portOut)
	portIn = func(data interface{}) {
		pd := getParseData(data)
		pos := pd.Source.pos
		n := len(pd.Source.content) - 1

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
		handleSemantics(portOut, portSemOut, setParseData, data, pd)
	}
	return
}

// ParseSpace parses one or more space characters.
// Space is defined by unicode.IsSpace().
// It can be configured wether EOL ('\n') is to be interpreted as space or not.
func ParseSpace(
	portOut func(interface{}),
	fillSemantics SemanticsOp,
	getParseData GetParseData,
	setParseData SetParseData,
	cfgEOLOK bool,
) (portIn func(interface{})) {
	portSemOut := makeSemanticsPort(fillSemantics, portOut)
	portIn = func(data interface{}) {
		var n int
		pd := getParseData(data)
		pos := pd.Source.pos
		substr := pd.Source.content[pos:]

		for i, char := range substr {
			if unicode.IsSpace(char) && (cfgEOLOK || char != '\n') {
				n = i + utf8.RuneLen(char)
			} else {
				break
			}
		}
		if n > 0 {
			createMatchedResult(pd, n)
		} else {
			createUnmatchedResult(pd, 0, "Expecting white space", nil)
		}
		handleSemantics(portOut, portSemOut, setParseData, data, pd)
	}
	return
}

// ParseRegexp parses text according to a predefined regular expression.
// The regular expression (e.g.: `^[a-z]+`) has to be configured.
// If the regular expression doesn't start with a `^` it will be added
// automatically.
// If the regular expression can't be compiled an error is returned.
func ParseRegexp(
	portOut func(interface{}),
	fillSemantics SemanticsOp,
	getParseData func(interface{}) *ParseData,
	setParseData func(interface{}, *ParseData) interface{},
	cfgRegexp string,
) (portIn func(interface{}), err error) {
	var re *regexp.Regexp
	if cfgRegexp[0] != '^' {
		cfgRegexp = "^" + cfgRegexp
	}
	if re, err = regexp.Compile(cfgRegexp); err != nil {
		return
	}

	portSemOut := makeSemanticsPort(fillSemantics, portOut)
	portIn = func(data interface{}) {
		pd := getParseData(data)
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
		handleSemantics(portOut, portSemOut, setParseData, data, pd)
	}
	return
}

// ParseLineComment parses a comment until the end of the line.
// The string that starts the comment (e.g.: `//`) has to be configured.
// If the start of the comment is empty an error is returned.
func ParseLineComment(
	portOut func(interface{}),
	fillSemantics SemanticsOp,
	getParseData func(interface{}) *ParseData,
	setParseData func(interface{}, *ParseData) interface{},
	cfgStart string,
) (portIn func(interface{}), err error) {
	if cfgStart == "" {
		return nil,
			errors.New(
				"expected start of line comment as config, got empty string",
			)
	}

	portSemOut := makeSemanticsPort(fillSemantics, portOut)
	portIn = func(data interface{}) {
		pd := getParseData(data)
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
		handleSemantics(portOut, portSemOut, setParseData, data, pd)
	}
	return
}

// ParseBlockComment parses a comment until the end of the line.
// The strings that start and end the comment (e.g.: `/*`, `*/`)
// have to be configured.
// A comment start or end inside a string literal (', " and `) is ignored.
// If the start or end of the comment is empty an error is returned.
func ParseBlockComment(
	portOut func(interface{}),
	fillSemantics SemanticsOp,
	getParseData func(interface{}) *ParseData,
	setParseData func(interface{}, *ParseData) interface{},
	cfgStart string,
	cfgEnd string,
) (portIn func(interface{}), err error) {
	if cfgStart == "" {
		return nil,
			errors.New(
				"expected start of block comment as config, got empty string",
			)
	}
	if cfgEnd == "" {
		return nil,
			errors.New(
				"expected end of block comment as config, got empty string",
			)
	}
	lBeg := len(cfgStart)
	lEnd := len(cfgEnd)

	portSemOut := makeSemanticsPort(fillSemantics, portOut)
	portIn = func(data interface{}) {
		pd := getParseData(data)
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
				case stringType == ' ':
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
		handleSemantics(portOut, portSemOut, setParseData, data, pd)
	}
	return
}
