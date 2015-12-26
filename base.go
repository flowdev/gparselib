package gparselib

import (
	"strconv"
	"strings"
)

//
// ---- Data types:

type Feedback struct {
	Errors   []*ParseError
	Warnings []string
	Infos    []string
}

type ParseResult struct {
	pos      int
	text     string
	value    interface{}
	errPos   int
	feedback Feedback
}

type SourceData struct {
	name        string
	content     string
	pos         int
	wherePrevNl int
	whereLine   int
}

func NewSourceData(name string, content string) *SourceData {
	return &SourceData{name, content, 0, -1, 1}
}

type ParseData struct {
	source     SourceData
	result     *ParseResult
	subResults []*ParseResult
}

func NewParseData(source SourceData) *ParseData {
	return &ParseData{source, nil, nil}
}

type ParseError struct {
	where   string
	myErr   string
	baseErr error
}

func (e *ParseError) Error() string {
	msg := "ERROR: " + e.where + e.myErr
	if e.baseErr != nil {
		msg += ":\n" + e.baseErr.Error()
	} else {
		msg += "."
	}
	return msg
}

// ------- Base for all parsers:

type Failable interface {
	SetErrorPort(func(error))
}
type SimpleParseOp interface {
	Failable
	InPort(interface{})
	SetOutPort(func(interface{}))
	SemInPort(interface{})
	SetSemOutPort(func(interface{}))
}

type BaseParseOp struct {
	parseData    func(interface{}) *ParseData
	setParseData func(interface{}, *ParseData) interface{}
	outPort      func(interface{})
	errorPort    func(error)
	semOutPort   func(interface{})
}

func (p *BaseParseOp) SemInPort(data interface{}) {
	p.outPort(data)
}
func (p *BaseParseOp) SetOutPort(outPort func(interface{})) {
	p.outPort = outPort
}
func (p *BaseParseOp) SetSemOutPort(semOutPort func(interface{})) {
	p.semOutPort = semOutPort
}
func (p *BaseParseOp) SetErrorPort(errPort func(error)) {
	p.errorPort = errPort
}
func (p *BaseParseOp) HandleSemantics(data interface{}, pd *ParseData) {
	if p.semOutPort != nil && pd.result.errPos < 0 {
		p.semOutPort(p.setParseData(data, pd))
	} else {
		p.outPort(p.setParseData(data, pd))
	}
}

//
// ---- Utility functions:

func createMatchedResult(pd *ParseData, n int) {
	i := pd.source.pos
	n += i
	pd.result = &ParseResult{i, pd.source.content[i:n], nil, -1, Feedback{}}
	pd.source.pos = n
}
func createUnmatchedResult(pd *ParseData, i int, msg string, baseErr error) {
	i += pd.source.pos
	pd.result = &ParseResult{pd.source.pos, "", nil, i, Feedback{}}
	pd.result.feedback.Errors = append(pd.result.feedback.Errors, &ParseError{where(&pd.source, i), msg, baseErr})
}
func where(src *SourceData, pos int) string {
	if pos >= src.wherePrevNl {
		return whereForward(src, pos)
	} else if pos <= src.wherePrevNl-pos {
		src.whereLine = 1
		src.wherePrevNl = -1
		return whereForward(src, pos)
	} else {
		return whereBackward(src, pos)
	}
}

func whereForward(src *SourceData, pos int) string {
	text := src.content
	lineNum := src.whereLine  // Line number
	prevNl := src.wherePrevNl // Line start (position of preceding newline)
	var nextNl int            // Position of next newline or end

	for {
		nextNl = strings.IndexByte(text[prevNl+1:], '\n')
		if nextNl < 0 {
			nextNl = len(text)
		} else {
			nextNl += prevNl + 1
		}
		where, stop := tryWhere(src, prevNl, pos, nextNl, lineNum)
		if stop {
			return where
		}
		prevNl = nextNl
		lineNum++
	}
}
func whereBackward(src *SourceData, pos int) string {
	text := src.content
	lineNum := src.whereLine  // Line number
	var prevNl int            // Line start (position of preceding newline)
	nextNl := src.wherePrevNl // Position of next newline or end

	for {
		prevNl = strings.LastIndexByte(text[0:nextNl], '\n')
		lineNum--
		where, stop := tryWhere(src, prevNl, pos, nextNl, lineNum)
		if stop {
			return where
		}
		nextNl = prevNl
	}
}
func tryWhere(src *SourceData, prevNl int, pos int, nextNl int, lineNum int) (where string, stop bool) {
	if prevNl < pos && pos <= nextNl {
		src.wherePrevNl = prevNl
		src.whereLine = lineNum
		return generateWhereMessage(src.name, lineNum, pos-prevNl, src.content[prevNl+1:nextNl]), true
	}
	return "", false
}
func generateWhereMessage(name string, line int, col int, srcLine string) string {
	return "File '" + name + "', line " + strconv.Itoa(line) +
		", column " + strconv.Itoa(col) + ":\n" + srcLine + "\n"
}

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}
