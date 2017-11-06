package gparselib

import (
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
) (
	portIn func(interface{}),
) {
	cfgN := len(cfgLiteral)

	portSemOut := makeSemanticsPort(fillSemantics, portOut)
	portIn = func(data interface{}) {
		pd := getParseData(data)
		pos := pd.Source.pos
		if len(pd.Source.content) >= pos+cfgN && pd.Source.content[pos:pos+cfgN] == cfgLiteral {
			createMatchedResult(pd, cfgN)
		} else {
			createUnmatchedResult(pd, 0, "Literal '"+cfgLiteral+"' expected", nil)
		}
		handleSemantics(portOut, portSemOut, setParseData, data, pd)
	}
	return
}

// This is needed for: ParseNatural
const allDigits = "0123456789abcdefghijklmnopqrstuvwxyz"

// ParseNatural parses a natural number at the current position of the parser.
// The configuration has to be the radix of accepted numbers (e.g.: 10).
func ParseNatural(
	portOut func(interface{}),
	fillSemantics SemanticsOp,
	getParseData GetParseData,
	setParseData SetParseData,
	cfgRadix int,
) (
	portIn func(interface{}),
	err error,
) {
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
) (
	portIn func(interface{}),
) {
	portSemOut := makeSemanticsPort(fillSemantics, portOut)
	portIn = func(data interface{}) {
		pd := getParseData(data)
		pos := pd.Source.pos
		n := len(pd.Source.content) - 1

		if n > pos {
			createUnmatchedResult(pd, 0,
				fmt.Sprintf(
					"Expecting end of input but got %d characters of input left",
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
) (
	portIn func(interface{}),
) {
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
// It can be configured wether EOL ('\n') is to be interpreted as space or not.
func ParseRegexp(
	portOut func(interface{}),
	fillSemantics SemanticsOp,
	getParseData func(interface{}) *ParseData,
	setParseData func(interface{}, *ParseData) interface{},
	cfgRegexp string,
) (
	portIn func(interface{}),
	err error,
) {
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
			createUnmatchedResult(pd, 0, "Expecting match for regexp `"+re.String()[1:]+"`", nil)
		}
		handleSemantics(portOut, portSemOut, setParseData, data, pd)
	}
	return
}

/*
// ------- Parse line comment:

type ParseLineComment struct {
	BaseParseOp
	start string
}

func NewParseLineComment(parseData func(interface{}) *ParseData,
	setParseData func(interface{}, *ParseData) interface{}, start string) *ParseLineComment {

	p := &ParseLineComment{}
	p.parseData = parseData
	p.setParseData = setParseData
	p.ConfigPort(start)
	return p
}
func (p *ParseLineComment) ConfigPort(start string) {
	p.start = start
}
func (p *ParseLineComment) InPort(data interface{}) {
	pd := p.parseData(data)
	pos := pd.Source.pos
	l := len(p.start)
	n := min(pos+l, len(pd.Source.content))
	substr := pd.Source.content[pos:n]

	if substr == p.start {
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
	p.HandleSemantics(data, pd)
}

// ------- Parse block comment:

type ParseBlockComment struct {
	BaseParseOp
	begin, end string
}

func NewParseBlockComment(parseData func(interface{}) *ParseData,
	setParseData func(interface{}, *ParseData) interface{}, begin, end string) *ParseBlockComment {

	p := &ParseBlockComment{}
	p.parseData = parseData
	p.setParseData = setParseData
	p.ConfigPort(begin, end)
	return p
}
func (p *ParseBlockComment) ConfigPort(begin, end string) {
	p.begin = begin
	p.end = end
}
func (p *ParseBlockComment) InPort(data interface{}) {
	pd := p.parseData(data)
	pos := pd.Source.pos
	lBeg := len(p.begin)
	lEnd := len(p.end)
	n := min(pos+lBeg, len(pd.Source.content))
	substr := pd.Source.content[pos:n]

	if substr == p.begin {
		afterBackslash := false
		stringType := ' '
		found := false
		endRune, _ := utf8.DecodeRuneInString(p.end)
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
				case '\'':
					stringType = '\''
				case '"':
					stringType = '"'
				case '`':
					stringType = '`'
				case endRune:
					if len(reststr) >= i+lEnd && reststr[i:i+lEnd] == p.end {
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
			createUnmatchedResult(pd, lBeg, "Block comment isn't closed properly", nil)
			pd.Source.pos += lBeg
		}
	} else {
		createUnmatchedResult(pd, 0, "Expecting block comment", nil)
	}
	p.HandleSemantics(data, pd)
}
*/
