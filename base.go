package gparselib

import (
	"fmt"
	"strconv"
	"strings"
)

//
// ---- Data types:

// FeedbackKind is just an enumeration of kinds of feedback.
type FeedbackKind int

// Enumeration of the kinds of feedback we can handle.
const (
	FeedbackInfo = FeedbackKind(iota)
	FeedbackWarning
	FeedbackError
)

// FeedbackItem is just one item of feedback.
type FeedbackItem struct {
	Kind FeedbackKind
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

// ParseResult contains the result of parsing including semantic value and
// feedback.
type ParseResult struct {
	Pos      int
	Text     string
	Value    interface{}
	ErrPos   int
	Feedback []*FeedbackItem
}

// HasError searches the feedback for errors and returns only true if it found
// one.
func (pr *ParseResult) HasError() bool {
	for _, fb := range pr.Feedback {
		if fb.Kind == FeedbackError {
			return true
		}
	}
	return false
}

// AddError adds a new parse error to the result feedback.
func (pd *ParseData) AddError(pos int, msg string, baseErr error) {
	pd.Result.Feedback = append(
		pd.Result.Feedback,
		&FeedbackItem{
			Kind: FeedbackError,
			Msg:  NewParseError(pd, pos, msg, baseErr),
		},
	)
}

// SourceData contains the name of the source for parsing, its contents and
// unexported stuff.
type SourceData struct {
	Name        string
	content     string
	pos         int
	wherePrevNl int
	whereLine   int
}

// NewSourceData creates a new, completely initialized SourceData.
func NewSourceData(name string, content string) SourceData {
	return SourceData{name, content, 0, -1, 1}
}

// Where describes the given integer position in a human readable way.
func (sd SourceData) Where(pos int) string {
	return where(&sd, pos)
}

// ParseData contains all data needed during parsing.
type ParseData struct {
	Source     SourceData
	Result     *ParseResult
	SubResults []*ParseResult
}

// NewParseData creates a new, completely initialized ParseData.
func NewParseData(name string, content string) *ParseData {
	return &ParseData{NewSourceData(name, content), nil, nil}
}

// ParseError holds information about a parser error.
type ParseError struct {
	where   string
	myErr   string
	baseErr error
}

// NewParseError creates a new, completely initialized ParseError.
func NewParseError(pd *ParseData, pos int, msg string, baseErr error) *ParseError {
	return &ParseError{where: pd.Source.Where(pos), myErr: msg, baseErr: baseErr}
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

// SemanticsOp is a simple filter for parser and context data.
type SemanticsOp func(pd *ParseData, ctx interface{}) (*ParseData, interface{})

// handleSemantics calls pluginSemantics if given and no error was detected, and always clears any subresults.
func handleSemantics(pluginSemantics SemanticsOp, pd *ParseData, ctx interface{}) (*ParseData, interface{}) {
	if pluginSemantics != nil && pd.Result.ErrPos < 0 {
		pd, ctx = pluginSemantics(pd, ctx)
	}
	pd.SubResults = nil
	return pd, ctx
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

// AddError adds an error feedback to the result part of the given ParseData.
func AddError(errPos int, msg string, baseErr error, pd *ParseData) {
	pd.Result.Feedback = append(
		pd.Result.Feedback,
		&FeedbackItem{
			Kind: FeedbackError,
			Msg:  NewParseError(pd, errPos, msg, baseErr),
		},
	)
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
