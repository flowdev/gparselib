package gparselib

import (
	"fmt"
	"math"
)

// SubparserOp is a simple filter to the outside and gets the same data as the
// parent parser.
type SubparserOp func(pd *ParseData, ctx interface{}) (*ParseData, interface{})

// ParseMulti uses a subparser multiple times.
// The minimum times the subparser has to match and the maximum times the
// subparser can match have to be configured.
func ParseMulti(
	pd *ParseData,
	ctx interface{},
	pluginSubparser SubparserOp,
	pluginSemantics SemanticsOp,
	cfgMin int,
	cfgMax int,
) (*ParseData, interface{}) {
	orgPos := pd.Source.pos
	relPos := 0
	subresults := make([]*ParseResult, 0, min(cfgMax, 128))

	for i := 0; i < cfgMax && pd.Result == nil; i++ {
		pd, ctx = pluginSubparser(pd, ctx)
		if !pd.Result.HasError() {
			subresults = append(subresults, pd.Result)
			pd.Result = nil
		}
	}

	relPos = pd.Source.pos - orgPos
	pd.Source.pos = orgPos
	if len(subresults) >= cfgMin {
		pd.Result = nil
		createMatchedResult(pd, relPos)
		saveAllValuesFeedback(pd, subresults)
	} else {
		subresult := pd.Result
		pd.Result = nil
		createUnmatchedResult(
			pd, relPos,
			fmt.Sprintf(
				"At least %d matches expected but got only %d",
				cfgMin,
				len(subresults),
			),
			nil,
		)
		pd.Result.Feedback = append(pd.Result.Feedback, subresult.Feedback...)
	}
	pd.SubResults = subresults
	return handleSemantics(pluginSemantics, pd, ctx)
}

// ParseMulti0 uses its subparser at least one time without upper bound.
// But the result is still positive even if the subparser didn't match a single
// time.
func ParseMulti0(
	pd *ParseData,
	ctx interface{},
	pluginSubparser SubparserOp,
	pluginSemantics SemanticsOp,
) (*ParseData, interface{}) {
	return ParseMulti(pd, ctx, pluginSubparser, pluginSemantics, 0, math.MaxInt32)
}

// ParseMulti1 uses its subparser at least one time without upper bound.
// The result is positive as long as the subparser matches at least one time.
func ParseMulti1(
	pd *ParseData,
	ctx interface{},
	pluginSubparser SubparserOp,
	pluginSemantics SemanticsOp,
) (*ParseData, interface{}) {
	return ParseMulti(pd, ctx, pluginSubparser, pluginSemantics, 1, math.MaxInt32)
}

// ParseOptional uses its subparser exaclty one time.
// But the result is still positive even if the subparser didn't match.
func ParseOptional(
	pd *ParseData,
	ctx interface{},
	pluginSubparser SubparserOp,
	pluginSemantics SemanticsOp,
) (*ParseData, interface{}) {
	orgPos := pd.Source.pos

	pd, ctx = pluginSubparser(pd, ctx)

	// if error: reset to ignore
	if pd.Result.HasError() {
		pd.Result.ErrPos = -1
		pd.Result.Feedback = nil
		pd.Source.pos = orgPos
	}

	return pd, ctx
}

// ParseAll calls multiple subparsers and all have to match for a successful result.
func ParseAll(
	pd *ParseData,
	ctx interface{},
	pluginSubparsers []SubparserOp,
	pluginSemantics SemanticsOp,
) (*ParseData, interface{}) {
	orgPos := pd.Source.pos
	subresults := make([]*ParseResult, len(pluginSubparsers))

	for i, subparser := range pluginSubparsers {
		pd, ctx = subparser(pd, ctx)
		if pd.Result.HasError() {
			pd.Source.pos = orgPos
			pd.Result.Pos = orgPos // make result 'our result'
			return pd, ctx
		}
		subresults[i] = pd.Result
		pd.Result = nil
	}

	relPos := pd.Source.pos - orgPos
	pd.Source.pos = orgPos
	createMatchedResult(pd, relPos)
	saveAllValuesFeedback(pd, subresults)
	pd.SubResults = subresults
	return handleSemantics(pluginSemantics, pd, ctx)
}

// ParseAny calls multiple subparsers until one matches.
// The result is only unsuccessful if no subparser matches.
func ParseAny(
	pd *ParseData,
	ctx interface{},
	pluginSubparsers []SubparserOp,
	pluginSemantics SemanticsOp,
) (*ParseData, interface{}) {
	orgPos := pd.Source.pos
	allFeedback := make([]*FeedbackItem, 0, len(pluginSubparsers))
	lastPos := 0

	for _, subparser := range pluginSubparsers {
		pd, ctx = subparser(pd, ctx)
		if !pd.Result.HasError() {
			return handleSemantics(pluginSemantics, pd, ctx)
		}
		lastPos = max(lastPos, pd.Result.Pos)
		pd.Source.pos = orgPos
		allFeedback = append(allFeedback, pd.Result.Feedback...)
		pd.Result = nil
	}

	relPos := lastPos - orgPos
	pd.Source.pos = orgPos
	pd.Result = nil
	createUnmatchedResult(
		pd, relPos,
		fmt.Sprintf(
			"Any subparser should match. But all %d subparsers failed",
			len(pluginSubparsers),
		),
		nil,
	)
	pd.Result.Feedback = append(pd.Result.Feedback, allFeedback...)
	return pd, ctx
}

// ParseBest calls all subparsers and chooses the one with the longest match.
// The result is only unsuccessful if no subparser matches.
func ParseBest(
	pd *ParseData,
	ctx interface{},
	pluginSubparsers []SubparserOp,
	pluginSemantics SemanticsOp,
) (*ParseData, interface{}) {
	orgPos := pd.Source.pos
	allFeedback := make([]*FeedbackItem, 0, len(pluginSubparsers))
	lastPos := 0
	bestNewSourcePos := 0
	var bestResult *ParseResult

	for _, subparser := range pluginSubparsers {
		pd, ctx = subparser(pd, ctx)
		if !pd.Result.HasError() {
			if pd.Source.pos > bestNewSourcePos {
				bestNewSourcePos = pd.Source.pos
				bestResult = pd.Result
			}
		}
		lastPos = max(lastPos, pd.Result.Pos)
		pd.Source.pos = orgPos
		allFeedback = append(allFeedback, pd.Result.Feedback...)
		pd.Result = nil
	}
	if bestResult != nil {
		pd.Source.pos = bestNewSourcePos
		pd.Result = bestResult
		return handleSemantics(pluginSemantics, pd, ctx)
	}

	relPos := lastPos - orgPos
	pd.Source.pos = orgPos
	pd.Result = nil
	createUnmatchedResult(
		pd, relPos,
		fmt.Sprintf(
			"Best subparser should match. But all %d subparsers failed",
			len(pluginSubparsers),
		),
		nil,
	)
	pd.Result.Feedback = append(pd.Result.Feedback, allFeedback...)
	return pd, ctx
}

func saveAllValuesFeedback(pd *ParseData, tmpSubresults []*ParseResult) {
	s := make([]interface{}, len(tmpSubresults))
	for i, subres := range tmpSubresults {
		s[i] = subres.Value
		pd.Result.Feedback = append(pd.Result.Feedback, subres.Feedback...)
	}
	pd.Result.Value = s
}