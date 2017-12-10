package gparselib

import (
	"fmt"
	"math"
)

// SubparserOp is a simple filter to the outside and gets the same data as the
// parent parser.
type SubparserOp func(portOut func(interface{})) (portIn func(interface{}))

// ParseMulti uses a subparser multiple times.
// The minimum times the subparser has to match and the maximum times the
// subparser can match have to be configured.
func ParseMulti(
	portOut func(interface{}),
	fillSubparser SubparserOp,
	fillSemantics SemanticsOp,
	getParseData GetParseData,
	setParseData SetParseData,
	cfgMin int,
	cfgMax int,
) (portIn func(interface{})) {
	defaultSemantics := func(data interface{}, pd *ParseData, tmp *tempData) {
		if cfgMax <= 1 {
			if len(tmp.subResults) >= 1 {
				pd.Result.Value = tmp.subResults[0].Value
				pd.Result.Feedback = append(pd.Result.Feedback, tmp.subResults[0].Feedback...)
			}
		} else {
			s := make([]interface{}, len(tmp.subResults))
			for i, subRes := range tmp.subResults {
				s[i] = subRes.Value
				pd.Result.Feedback = append(pd.Result.Feedback, subRes.Feedback...)
			}
			pd.Result.Value = s
		}
	}

	var portSubIn func(interface{})
	portSemOut := makeSemanticsPort(fillSemantics, portOut)
	handleSubresult := func(data interface{}) {
		pd := getParseData(data)
		tmp := pd.tmp[len(pd.tmp)-1]
		i := len(tmp.subResults)
		if pd.Result.ErrPos >= 0 {
			relPos := pd.Result.Pos - tmp.pos
			pd.Source.pos = tmp.pos
			if i >= cfgMin {
				pd.Result = nil
				createMatchedResult(pd, relPos)
				defaultSemantics(data, pd, tmp)
			} else {
				subResult := pd.Result
				pd.Result = nil
				createUnmatchedResult(pd, relPos, fmt.Sprintf("At least %d matches expected but got only %d.", cfgMin, i), nil)
				pd.Result.Feedback = append(pd.Result.Feedback, subResult.Feedback...)
			}
			pd.SubResults = tmp.subResults
			pd.tmp = pd.tmp[:len(pd.tmp)-1]
			handleSemantics(portOut, portSemOut, setParseData, data, pd)
		} else {
			tmp.subResults = append(tmp.subResults, pd.Result)
			pd.Result = nil
			if i+1 >= cfgMax {
				relPos := pd.Source.pos - tmp.pos
				pd.Source.pos = tmp.pos
				createMatchedResult(pd, relPos)
				defaultSemantics(data, pd, tmp)
				pd.SubResults = tmp.subResults
				pd.tmp = pd.tmp[:len(pd.tmp)-1]
				handleSemantics(portOut, portSemOut, setParseData, data, pd)
			} else {
				portSubIn(data)
			}
		}
	}
	portSubIn = fillSubparser(handleSubresult)
	portIn = func(data interface{}) {
		pd := getParseData(data)
		tmp := newTempData(pd.Source.pos, min(cfgMax, 128))
		pd.tmp = append(pd.tmp, tmp)
		portSubIn(setParseData(data, pd))
	}
	return
}

// ParseMulti0 uses its subparser at least one time without upper bound.
// But the result is still positive even if the subparser didn't match a single
// time.
func ParseMulti0(
	portOut func(interface{}),
	fillSubparser SubparserOp,
	fillSemantics SemanticsOp,
	getParseData GetParseData,
	setParseData SetParseData,
) (portIn func(interface{})) {
	return ParseMulti(
		portOut,
		fillSubparser,
		fillSemantics,
		getParseData,
		setParseData,
		0,
		math.MaxInt32,
	)
}

// ParseMulti1 uses its subparser at least one time without upper bound.
// The result is positive as long as the subparser matches at least one time.
func ParseMulti1(
	portOut func(interface{}),
	fillSubparser SubparserOp,
	fillSemantics SemanticsOp,
	getParseData GetParseData,
	setParseData SetParseData,
) (portIn func(interface{})) {
	return ParseMulti(
		portOut,
		fillSubparser,
		fillSemantics,
		getParseData,
		setParseData,
		1,
		math.MaxInt32,
	)
}

// ParseOptional uses its subparser exaclty one time.
// But the result is still positive even if the subparser didn't match.
func ParseOptional(
	portOut func(interface{}),
	fillSubparser SubparserOp,
	fillSemantics SemanticsOp,
	getParseData GetParseData,
	setParseData SetParseData,
) (portIn func(interface{})) {
	return ParseMulti(
		portOut,
		fillSubparser,
		fillSemantics,
		getParseData,
		setParseData,
		0,
		1,
	)
}

// ParseAll calls multiple subparsers and all have to match for a successful result.
func ParseAll(
	portOut func(interface{}),
	fillSubparsers []SubparserOp,
	fillSemantics SemanticsOp,
	getParseData GetParseData,
	setParseData SetParseData,
) (portIn func(interface{})) {
	defaultSemantics := func(pd *ParseData, tmp *tempData) {
		s := make([]interface{}, len(tmp.subResults))
		for i, subRes := range tmp.subResults {
			s[i] = subRes.Value
			pd.Result.Feedback = append(pd.Result.Feedback, subRes.Feedback...)
		}
		pd.Result.Value = s
	}

	handleSubresult := func(data interface{}, pd *ParseData, tmp *tempData) *ParseResult {
		switch {
		case pd.Result.ErrPos < 0 && len(tmp.subResults)+1 >= len(fillSubparsers):
			tmp.subResults = append(tmp.subResults, pd.Result)
			pd.Result = nil
			relPos := pd.Source.pos - tmp.pos
			pd.Source.pos = tmp.pos
			createMatchedResult(pd, relPos)
			defaultSemantics(pd, tmp)
			return nil
		case pd.Result.ErrPos < 0:
			return pd.Result
		default:
			// pd.Result is set by subparser!
			pd.Result.Pos = tmp.pos // but we have to correct the position to the overall result position
			pd.Source.pos = tmp.pos
			return nil
		}
	}

	portIn = parseWithMultiSubOps(
		portOut,
		fillSubparsers,
		fillSemantics,
		getParseData,
		setParseData,
		handleSubresult,
	)
	return
}

// ParseAny calls multiple subparsers until one matches.
// The result is only unsuccessful if no subparser matches.
func ParseAny(
	portOut func(interface{}),
	fillSubparsers []SubparserOp,
	fillSemantics SemanticsOp,
	getParseData GetParseData,
	setParseData SetParseData,
) (portIn func(interface{})) {
	defaultSemantics := func(pd *ParseData, tmp *tempData) {
		subRes := tmp.subResults[len(tmp.subResults)-1]
		pd.Result.Value = subRes.Value
		pd.Result.Feedback = append(pd.Result.Feedback, subRes.Feedback...)
	}

	handleSubresult := func(data interface{}, pd *ParseData, tmp *tempData) *ParseResult {
		switch {
		case pd.Result.ErrPos < 0:
			tmp.subResults = append(tmp.subResults, pd.Result)
			pd.Result = nil
			relPos := pd.Source.pos - tmp.pos
			pd.Source.pos = tmp.pos
			createMatchedResult(pd, relPos)
			defaultSemantics(pd, tmp)
			return nil
		case pd.Result.ErrPos >= 0 && len(tmp.subResults)+1 >= len(fillSubparsers):
			tmp.subResults = append(tmp.subResults, pd.Result)
			relPos := pd.Result.Pos - tmp.pos
			pd.Source.pos = tmp.pos
			pd.Result = nil
			createUnmatchedResult(pd, relPos, fmt.Sprintf("Any subparser should match. But all %d subparsers failed", len(fillSubparsers)), nil)
			for _, subRes := range tmp.subResults {
				pd.Result.Feedback = append(pd.Result.Feedback, subRes.Feedback...)
			}
			return nil
		default:
			pd.Source.pos = tmp.pos
			return pd.Result
		}
	}

	portIn = parseWithMultiSubOps(
		portOut,
		fillSubparsers,
		fillSemantics,
		getParseData,
		setParseData,
		handleSubresult,
	)
	return
}

// parseWithMultiSubOps is a base operation that is used in parser operations that have multiple subparsers.
// The contract is the following:
// parseWithMultiSubOps:
//	- handle pd.tmp completely (so please don't touch)
//	- fill tmp.pos
//	and finally
//	- call fillSemantics
//
// The parser op:
//	- handleSubresult must create the matched or unmatched Result and must return nil if parsing should stop
//	- handleSubresult *must* return a non-nil subResult if parsing should go on
//	- may do some default semantics in handleSubresult
func parseWithMultiSubOps(
	portOut func(interface{}),
	fillSubparsers []SubparserOp,
	fillSemantics SemanticsOp,
	getParseData GetParseData,
	setParseData SetParseData,
	handleSubresult func(interface{}, *ParseData, *tempData) *ParseResult,
) (portIn func(interface{})) {
	portSubIns := make([]func(interface{}), len(fillSubparsers)) // forward declaration
	portSemOut := makeSemanticsPort(fillSemantics, portOut)

	portMySubIn := func(data interface{}) { // call subparsers as long as: subResult != nil
		pd := getParseData(data)
		tmp := pd.tmp[len(pd.tmp)-1]
		subResult := handleSubresult(data, pd, tmp)
		if subResult == nil {
			pd.SubResults = tmp.subResults
			pd.tmp = pd.tmp[0 : len(pd.tmp)-1]
			handleSemantics(portOut, portSemOut, setParseData, data, pd)
		} else {
			tmp.subResults = append(tmp.subResults, subResult)
			pd.Result = nil
			portSubIns[len(tmp.subResults)](data)
		}
	}

	for i, fillSubparser := range fillSubparsers { // convert fills to ports
		portSubIns[i] = fillSubparser(portMySubIn)
	}

	portIn = func(data interface{}) {
		pd := getParseData(data)
		tmp := newTempData(pd.Source.pos, len(fillSubparsers))
		pd.tmp = append(pd.tmp, tmp)
		portSubIns[0](data)
	}
	return
}
