package gparselib

import (
	"errors"
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
	FeedbackUnknown = FeedbackKind(iota) // should never be used directly but is default value
	FeedbackInfo
	FeedbackWarning
	FeedbackPotentialProblem
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
	case FeedbackPotentialProblem:
		msg = "PROBLEM?: "
	case FeedbackError:
		msg = "ERROR: "
	default:
		msg = "UNKNOWN!!!: "
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

// AddInfo adds a new parse information to the result feedback.
func (pd *ParseData) AddInfo(pos int, msg string) {
	pd.Result.Feedback = append(
		pd.Result.Feedback,
		&FeedbackItem{
			Kind: FeedbackInfo,
			Msg:  newParseMessage(pd, pos, msg),
		},
	)
}

// AddWarning adds a new parser warning to the result feedback.
func (pd *ParseData) AddWarning(pos int, msg string) {
	pd.Result.Feedback = append(
		pd.Result.Feedback,
		&FeedbackItem{
			Kind: FeedbackWarning,
			Msg:  newParseMessage(pd, pos, msg),
		},
	)
}

// AddError adds a new parse error to the result feedback.
func (pd *ParseData) AddError(pos int, msg string, baseErr error) {
	pd.Result.Feedback = append(
		pd.Result.Feedback,
		&FeedbackItem{
			Kind: FeedbackError,
			Msg:  newParseError(pd, pos, msg, baseErr),
		},
	)
}

// GetFeedback returns parser errors as a single error and
// additional feedback.
func (pd *ParseData) GetFeedback() (string, error) {
	return feedbackInfo(pd.Result), feedbackError(pd.Result)
}
func feedbackError(pr *ParseResult) error {
	b := strings.Builder{}
	for _, fb := range pr.Feedback {
		if fb.Kind == FeedbackError {
			if b.Len() > 0 {
				b.WriteString("\n")
			}
			b.WriteString(fb.String())
		}
	}
	if b.Len() == 0 {
		return nil
	}
	return errors.New(b.String())
}
func feedbackInfo(pr *ParseResult) string {
	b := strings.Builder{}
	for _, fb := range pr.Feedback {
		if fb.Kind != FeedbackError {
			if b.Len() > 0 {
				b.WriteString("\n")
			}
			b.WriteString(fb.String())
		}
	}
	return b.String()
}

// ResetSourcePos resets the source position to an old value.
// This is usually needed when semantic errors occur.
// If the given pos is negative, the position of the current result is used.
func (pd *ParseData) ResetSourcePos(pos int) {
	if pos < 0 {
		pd.Source.pos = pd.Result.Pos
	} else {
		pd.Source.pos = pos
	}
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

// Where describes the given integer position in a human-readable way.
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
	return &ParseData{Source: NewSourceData(name, content)}
}

// parseMessage holds some information from the parser.
type parseMessage struct {
	where string
	msg   string
}

// newParseMessage creates a new, completely initialized parseMessage.
func newParseMessage(pd *ParseData, pos int, msg string) *parseMessage {
	return &parseMessage{where: pd.Source.Where(pos), msg: msg}
}
func (i *parseMessage) String() string {
	b := &strings.Builder{}
	b.WriteString(i.where)
	b.WriteString(i.msg)
	b.WriteRune('.')

	return b.String()
}

// parseError holds information about a parser error.
type parseError struct {
	where   string
	myErr   string
	baseErr error
}

// newParseError creates a new, completely initialized parseError.
func newParseError(pd *ParseData, pos int, msg string, baseErr error) *parseError {
	return &parseError{where: pd.Source.Where(pos), myErr: msg, baseErr: baseErr}
}

func (e *parseError) Error() string {
	msg := e.where + e.myErr
	if e.baseErr != nil {
		msg += ":\n" + e.baseErr.Error()
	} else {
		msg += "."
	}
	return msg
}
func (e *parseError) String() string {
	return e.Error()
}

// ------- Base for all parsers:

// SubparserOp is a simple filter to the outside and gets the same data as the
// parent parser.
type SubparserOp func(pd *ParseData, ctx interface{}) (*ParseData, interface{})

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
	pd.AddError(i, msg, baseErr)
}

func where(src *SourceData, pos int) string {
	if src.content == "" {
		return generateWhereMessage(src.Name, 1, 1, "")
	}
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
