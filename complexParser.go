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
