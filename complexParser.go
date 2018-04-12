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
		if pd.Result.ErrPos < 0 {
			subresults = append(subresults, pd.Result)
			pd.Result = nil
		}
	}

	relPos = pd.Source.pos - orgPos
	pd.Source.pos = orgPos
	if len(subresults) >= cfgMin {
		pd.Result = nil
		createMatchedResult(pd, relPos)
		parseMultiDefaultSemantics(pd, subresults, cfgMax <= 1)
	} else {
		subResult := pd.Result
		pd.Result = nil
		createUnmatchedResult(pd, relPos, fmt.Sprintf("At least %d matches expected but got only %d.", cfgMin, len(subresults)), nil)
		pd.Result.Feedback = append(pd.Result.Feedback, subResult.Feedback...)
	}
	pd.SubResults = subresults
	return handleSemantics(pluginSemantics, pd, ctx)
}
func parseMultiDefaultSemantics(pd *ParseData, tmpSubresults []*ParseResult, singleResult bool) {
	if singleResult {
		saveSingleValueFeedback(pd, tmpSubresults)
	} else {
		saveAllValuesFeedback(pd, tmpSubresults)
	}
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
	return ParseMulti(pd, ctx, pluginSubparser, pluginSemantics, 0, 1)
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
		if pd.Result.ErrPos >= 0 {
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
		if pd.Result.ErrPos < 0 {
			relPos := pd.Source.pos - orgPos
			pd.Source.pos = orgPos
			createMatchedResult(pd, relPos)
			return handleSemantics(pluginSemantics, pd, ctx)
		}
		lastPos = pd.Result.Pos
		pd.Source.pos = orgPos
		allFeedback = append(allFeedback, pd.Result.Feedback...)
		pd.Result = nil
	}

	relPos := lastPos - orgPos
	pd.Source.pos = orgPos
	pd.Result = nil
	createUnmatchedResult(pd, relPos, fmt.Sprintf("Any subparser should match. But all %d subparsers failed", len(pluginSubparsers)), nil)
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
func saveSingleValueFeedback(pd *ParseData, tmpSubresults []*ParseResult) {
	if len(tmpSubresults) >= 1 {
		subRes := tmpSubresults[0]
		pd.Result.Value = subRes.Value
		pd.Result.Feedback = append(pd.Result.Feedback, subRes.Feedback...)
	}
}
