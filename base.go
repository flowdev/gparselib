package gparselib

import (
	"fmt"
	"strconv"
	"strings"
)

//
// ---- Data types:

const (
	FeedbackInfo = iota
	FeedbackWarning
	FeedbackError
)

type FeedbackItem struct {
	Kind int
	Msg  fmt.Stringer
}

func (fi *FeedbackItem) String() string {
	var msg string
	switch fi.Kind {
	case FeedbackInfo:
		msg = "INFO: "
	case FeedbackWarning:
		msg = "WARNING: "
	case FeedbackError:
		msg = "ERROR: "
	}
	return msg + fi.Msg.String()
}

type ParseResult struct {
	Pos      int
	Text     string
	Value    interface{}
	ErrPos   int
	Feedback []*FeedbackItem
}

func (pr *ParseResult) HasError() bool {
	for _, fb := range pr.Feedback {
		if fb.Kind == FeedbackError {
			return true
		}
	}
	return false
}

type SourceData struct {
	Name        string
	content     string
	pos         int
	wherePrevNl int
	whereLine   int
}

func NewSourceData(name string, content string) SourceData {
	return SourceData{name, content, 0, -1, 1}
}

type tempData struct {
	pos        int
	subResults []*ParseResult
}

func newTempData(pos, n int) *tempData {
	return &tempData{pos, make([]*ParseResult, 0, n)}
}

type ParseData struct {
	Source     SourceData
	Result     *ParseResult
	SubResults []*ParseResult
	tmp        []*tempData
}

func NewParseData(name string, content string) *ParseData {
	return &ParseData{NewSourceData(name, content), nil, nil, make([]*tempData, 0, 128)}
}

type ParseError struct {
	where   string
	myErr   string
	baseErr error
}

func NewParseError(pd *ParseData, pos int, msg string, baseErr error) *ParseError {
	return &ParseError{where(&pd.Source, pos), msg, baseErr}
}
func (e *ParseError) Error() string {
	msg := e.where + e.myErr
	if e.baseErr != nil {
		msg += ":\n" + e.baseErr.Error()
	} else {
		msg += "."
	}
	return msg
}
func (e *ParseError) String() string {
	return e.Error()
}

// ------- Base for all parsers:

type SimpleParseOp interface {
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

func (p *BaseParseOp) SemInPort(dat interface{}) {
	pd := p.parseData(dat)
	pd.SubResults = nil
	p.outPort(dat)
}
func (p *BaseParseOp) SetOutPort(outPort func(interface{})) {
	p.outPort = outPort
}
func (p *BaseParseOp) SetSemOutPort(semOutPort func(interface{})) {
	p.semOutPort = semOutPort
}
func (p *BaseParseOp) HandleSemantics(dat interface{}, pd *ParseData) {
	if p.semOutPort != nil && pd.Result.ErrPos < 0 {
		p.semOutPort(p.setParseData(dat, pd))
	} else {
		pd.SubResults = nil
		p.outPort(p.setParseData(dat, pd))
	}
}

//
// ---- Utility functions:

func createMatchedResult(pd *ParseData, n int) {
	i := pd.Source.pos
	n += i
	pd.Result = &ParseResult{i, pd.Source.content[i:n], nil, -1, make([]*FeedbackItem, 0, 64)}
	pd.Source.pos = n
}
func createUnmatchedResult(pd *ParseData, i int, msg string, baseErr error) {
	i += pd.Source.pos
	pd.Result = &ParseResult{pd.Source.pos, "", nil, i, make([]*FeedbackItem, 0, 64)}
	AddError(i, msg, baseErr, pd)
}

func AddError(errPos int, msg string, baseErr error, pd *ParseData) {
	pd.Result.Feedback = append(pd.Result.Feedback, &FeedbackItem{Kind: FeedbackError, Msg: NewParseError(pd, errPos, msg, baseErr)})
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
		return generateWhereMessage(src.Name, lineNum, pos-prevNl, src.content[prevNl+1:nextNl]), true
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
func max(a, b int) int {
	if a >= b {
		return a
	}
	return b
}
