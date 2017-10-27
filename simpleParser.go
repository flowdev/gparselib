package gparselib

/*
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

// ------- Parse regexp:

type ParseRegexp struct {
	BaseParseOp
	re *regexp.Regexp
}

func NewParseRegexp(parseData func(interface{}) *ParseData,
	setParseData func(interface{}, *ParseData) interface{}, re string) *ParseRegexp {

	p := &ParseRegexp{}
	p.parseData = parseData
	p.setParseData = setParseData
	p.ConfigPort(re)
	return p
}
func (p *ParseRegexp) ConfigPort(s string) {
	var t string
	if s[0] == '^' {
		t = s
	} else {
		t = "^" + s
	}
	re, err := regexp.Compile(t)
	if err != nil {
		p.errorPort(err)
	} else {
		p.re = re
	}
}
func (p *ParseRegexp) InPort(data interface{}) {
	pd := p.parseData(data)
	pos := pd.Source.pos
	substr := pd.Source.content[pos:]
	match := p.re.FindStringIndex(substr)

	if match != nil {
		createMatchedResult(pd, match[1])
		pd.Result.Value = pd.Result.Text
	} else {
		createUnmatchedResult(pd, 0, "Expecting match for regexp `"+p.re.String()[1:]+"`", nil)
	}
	p.HandleSemantics(data, pd)
}

// ------- Parse space:

type ParseSpace struct {
	BaseParseOp
	eolOk bool
}

func NewParseSpace(parseData func(interface{}) *ParseData,
	setParseData func(interface{}, *ParseData) interface{}, eolOk bool) *ParseSpace {

	p := &ParseSpace{}
	p.parseData = parseData
	p.setParseData = setParseData
	p.ConfigPort(eolOk)
	return p
}
func (p *ParseSpace) ConfigPort(eolOk bool) {
	p.eolOk = eolOk
}
func (p *ParseSpace) InPort(data interface{}) {
	var n int
	pd := p.parseData(data)
	pos := pd.Source.pos
	substr := pd.Source.content[pos:]

	for i, char := range substr {
		if unicode.IsSpace(char) && (p.eolOk || char != '\n') {
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
	p.HandleSemantics(data, pd)
}

// ------- Parse EOF:

type ParseEof struct {
	BaseParseOp
}

func NewParseEof(parseData func(interface{}) *ParseData,
	setParseData func(interface{}, *ParseData) interface{}) *ParseEof {

	p := &ParseEof{}
	p.parseData = parseData
	p.setParseData = setParseData
	return p
}
func (p *ParseEof) InPort(data interface{}) {
	pd := p.parseData(data)
	pos := pd.Source.pos
	n := len(pd.Source.content) - 1

	if n > pos {
		createUnmatchedResult(pd, 0, "Expecting end of input but got "+strconv.Itoa(n-pos)+
			" characters of input left", nil)
	} else {
		createMatchedResult(pd, 0)
	}
	p.HandleSemantics(data, pd)
}

// ------- Parse a natural number:

const allDigits = "0123456789abcdefghijklmnopqrstuvwxyz"

type ParseNatural struct {
	BaseParseOp
	cfgDigits string
}

func NewParseNatural(parseData func(interface{}) *ParseData,
	setParseData func(interface{}, *ParseData) interface{},
	radix int) *ParseNatural {

	p := &ParseNatural{}
	p.parseData = parseData
	p.setParseData = setParseData
	p.ConfigPort(radix)
	return p
}
func (p *ParseNatural) ConfigPort(radix int) {
	// panic if radix < 2 or radix > 36 !!!
	if radix < 2 || radix > 36 {
		panic(&ParseError{"", "The given radix of '" + strconv.Itoa(radix) + "' is out of the allowed range from 2 to 36.", nil})
	}
	p.cfgDigits = allDigits[0:radix]
}
func (p *ParseNatural) InPort(data interface{}) {
	var n int
	pd := p.parseData(data)
	pos := pd.Source.pos
	substr := pd.Source.content[pos:]

	for i, digit := range substr {
		if strings.IndexRune(p.cfgDigits, unicode.ToLower(digit)) >= 0 {
			n = i + 1
		} else {
			break
		}
	}
	if n > 0 {
		val, err := strconv.ParseUint(substr[0:n], len(p.cfgDigits), 64)
		if err == nil {
			createMatchedResult(pd, n)
			pd.Result.Value = val
		} else {
			createUnmatchedResult(pd, 0, "Natural number expected", err)
		}
	} else {
		createUnmatchedResult(pd, 0, "Natural number expected", nil)
	}
	p.HandleSemantics(data, pd)
}
*/

// ------- Parse a literal value:

// ParseLiteral parses a literal value at the current position of the parser.
// The configuration has to be the literal string we expect.
func ParseLiteral(
	outPort func(interface{}),
	semOutPort func(interface{}),
	getParseData func(interface{}) *ParseData,
	setParseData func(interface{}, *ParseData) interface{},
	cfgLiteral string,
) (
	inPort func(interface{}),
	semInPort func(interface{}),
	setSemOutPort func(func(interface{})),
) {
	cfgN := len(cfgLiteral)
	setSemOutPort = func(sop func(interface{})) { semOutPort = sop }
	semInPort = defaultSemInPort(outPort, getParseData)
	inPort = func(data interface{}) {
		pd := getParseData(data)
		pos := pd.Source.pos
		if len(pd.Source.content) >= pos+cfgN && pd.Source.content[pos:pos+cfgN] == cfgLiteral {
			createMatchedResult(pd, cfgN)
		} else {
			createUnmatchedResult(pd, 0, "Literal '"+cfgLiteral+"' expected", nil)
		}
		handleSemantics(outPort, semOutPort, setParseData, data, pd)
	}
	return
}
