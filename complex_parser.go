package gparselib

import (
	"fmt"
	"math"
)

// ParseMulti uses a subparser multiple times.
// The minimum times the subparser has to match and the maximum times the
// subparser can match have to be configured.
func ParseMulti(
	pd *ParseData, ctx interface{},
	pluginSubparser SubparserOp, pluginSemantics SemanticsOp,
	cfgMin, cfgMax int,
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

// NewParseMultiPlugin creates a plugin sporting a parser calling a subparser
// multiple times.
func NewParseMultiPlugin(
	pluginSubparser SubparserOp, pluginSemantics SemanticsOp,
	cfgMin, cfgMax int,
) SubparserOp {
	return func(pd *ParseData, ctx interface{}) (*ParseData, interface{}) {
		return ParseMulti(pd, ctx, pluginSubparser, pluginSemantics, cfgMin, cfgMax)
	}
}

// ParseMulti0 uses its subparser at least one time without upper bound.
// But the result is still positive even if the subparser didn't match a single
// time.
func ParseMulti0(
	pd *ParseData, ctx interface{},
	pluginSubparser SubparserOp, pluginSemantics SemanticsOp,
) (*ParseData, interface{}) {
	return ParseMulti(pd, ctx, pluginSubparser, pluginSemantics, 0, math.MaxInt32)
}

// NewParseMulti0Plugin creates a plugin sporting a parser calling a subparser
// multiple times.
func NewParseMulti0Plugin(pluginSubparser SubparserOp, pluginSemantics SemanticsOp) SubparserOp {
	return func(pd *ParseData, ctx interface{}) (*ParseData, interface{}) {
		return ParseMulti0(pd, ctx, pluginSubparser, pluginSemantics)
	}
}

// ParseMulti1 uses its subparser at least one time without upper bound.
// The result is positive as long as the subparser matches at least one time.
func ParseMulti1(
	pd *ParseData, ctx interface{},
	pluginSubparser SubparserOp, pluginSemantics SemanticsOp,
) (*ParseData, interface{}) {
	return ParseMulti(pd, ctx, pluginSubparser, pluginSemantics, 1, math.MaxInt32)
}

// NewParseMulti1Plugin creates a plugin sporting a parser calling a subparser
// at least one time.
func NewParseMulti1Plugin(pluginSubparser SubparserOp, pluginSemantics SemanticsOp) SubparserOp {
	return func(pd *ParseData, ctx interface{}) (*ParseData, interface{}) {
		return ParseMulti1(pd, ctx, pluginSubparser, pluginSemantics)
	}
}

// ParseOptional uses its subparser exaclty one time.
// But the result is still positive even if the subparser didn't match.
func ParseOptional(
	pd *ParseData, ctx interface{},
	pluginSubparser SubparserOp, pluginSemantics SemanticsOp,
) (*ParseData, interface{}) {
	orgPos := pd.Source.pos

	pd, ctx = pluginSubparser(pd, ctx)

	// if error: reset to ignore
	if pd.Result.HasError() {
		pd.Result.ErrPos = -1
		pd.Result.Feedback = nil
		pd.Source.pos = orgPos
	}

	return handleSemantics(pluginSemantics, pd, ctx)
}

// NewParseOptionalPlugin creates a plugin sporting a parser calling a subparser
// once and ignoring errors in it.
func NewParseOptionalPlugin(pluginSubparser SubparserOp, pluginSemantics SemanticsOp) SubparserOp {
	return func(pd *ParseData, ctx interface{}) (*ParseData, interface{}) {
		return ParseOptional(pd, ctx, pluginSubparser, pluginSemantics)
	}
}

// ParseAll calls multiple subparsers and all have to match for a successful result.
func ParseAll(
	pd *ParseData, ctx interface{},
	pluginSubparsers []SubparserOp, pluginSemantics SemanticsOp,
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

// NewParseAllPlugin creates a plugin sporting a parser calling all subparsers.
func NewParseAllPlugin(
	pluginSubparsers []SubparserOp, pluginSemantics SemanticsOp,
) SubparserOp {
	return func(pd *ParseData, ctx interface{}) (*ParseData, interface{}) {
		return ParseAll(pd, ctx, pluginSubparsers, pluginSemantics)
	}
}

// ParseAny calls multiple subparsers until one matches.
// The result is only unsuccessful if no subparser matches.
func ParseAny(
	pd *ParseData, ctx interface{},
	pluginSubparsers []SubparserOp, pluginSemantics SemanticsOp,
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

// NewParseAnyPlugin creates a plugin sporting a parser calling any successful
// subparser.
func NewParseAnyPlugin(
	pluginSubparsers []SubparserOp, pluginSemantics SemanticsOp,
) SubparserOp {
	return func(pd *ParseData, ctx interface{}) (*ParseData, interface{}) {
		return ParseAny(pd, ctx, pluginSubparsers, pluginSemantics)
	}
}

// ParseBest calls all subparsers and chooses the one with the longest match.
// The result is only unsuccessful if no subparser matches.
func ParseBest(
	pd *ParseData, ctx interface{},
	pluginSubparsers []SubparserOp, pluginSemantics SemanticsOp,
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

// NewParseBestPlugin creates a plugin sporting a parser calling the best
// successful subparser.
func NewParseBestPlugin(
	pluginSubparsers []SubparserOp, pluginSemantics SemanticsOp,
) SubparserOp {
	return func(pd *ParseData, ctx interface{}) (*ParseData, interface{}) {
		return ParseBest(pd, ctx, pluginSubparsers, pluginSemantics)
	}
}

func saveAllValuesFeedback(pd *ParseData, tmpSubresults []*ParseResult) {
	s := make([]interface{}, len(tmpSubresults))
	for i, subres := range tmpSubresults {
		s[i] = subres.Value
		pd.Result.Feedback = append(pd.Result.Feedback, subres.Feedback...)
	}
	pd.Result.Value = s
}
