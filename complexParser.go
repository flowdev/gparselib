package gparselib

import (
	"fmt"
	"math"
)

// ------- Parse multiple subparsers with configurable min and max count:
type ParseMulti struct {
	BaseParseOp
	subOutPort func(interface{})
	min, max   int
}
type ParseMultiConfig struct {
	Min, Max int
}

func NewParseMulti(parseData func(interface{}) *ParseData,
	setParseData func(interface{}, *ParseData) interface{},
	min, max int) *ParseMulti {

	p := &ParseMulti{}
	p.parseData = parseData
	p.setParseData = setParseData
	p.ConfigPort(ParseMultiConfig{min, max})
	return p
}
func (p *ParseMulti) ConfigPort(cfg ParseMultiConfig) {
	p.min, p.max = cfg.Min, cfg.Max
}
func (p *ParseMulti) SetSubOutPort(subOutPort func(interface{})) {
	p.subOutPort = subOutPort
}
func (p *ParseMulti) SubInPort(data interface{}) {
	pd := p.parseData(data)
	p.handleSubresult(data, pd, pd.tmp[len(pd.tmp)-1])
}
func (p *ParseMulti) InPort(data interface{}) {
	pd := p.parseData(data)
	tmp := newTempData(pd.source.pos, min(p.max, 128))
	pd.tmp = append(pd.tmp, tmp)
	p.subOutPort(data)
}
func (p *ParseMulti) handleSubresult(data interface{}, pd *ParseData, tmp *tempData) {
	i := len(tmp.subResults)
	if pd.Result.ErrPos >= 0 {
		relPos := pd.Result.Pos - tmp.pos
		pd.source.pos = tmp.pos
		if i >= p.min {
			pd.Result = nil
			createMatchedResult(pd, relPos)
			p.defaultSemantics(data, pd, tmp)
		} else {
			subResult := pd.Result
			pd.Result = nil
			createUnmatchedResult(pd, relPos, fmt.Sprintf("At least %d matches expected but got only %d.", p.min, i), nil)
			AddFeedback(&(pd.Result.Feedback), subResult.Feedback)
		}
		pd.SubResults = tmp.subResults
		pd.tmp = pd.tmp[0 : len(pd.tmp)-1]
		p.HandleSemantics(data, pd)
	} else {
		tmp.subResults = append(tmp.subResults, pd.Result)
		pd.Result = nil
		if i+1 >= p.max {
			relPos := pd.source.pos - tmp.pos
			pd.source.pos = tmp.pos
			createMatchedResult(pd, relPos)
			p.defaultSemantics(data, pd, tmp)
			pd.SubResults = tmp.subResults
			pd.tmp = pd.tmp[0 : len(pd.tmp)-1]
			p.HandleSemantics(data, pd)
		} else {
			p.subOutPort(data)
		}
	}
}
func (p *ParseMulti) defaultSemantics(data interface{}, pd *ParseData, tmp *tempData) {
	if p.max <= 1 {
		if len(tmp.subResults) >= 1 {
			pd.Result.Value = tmp.subResults[0].Value
			AddFeedback(&(pd.Result.Feedback), tmp.subResults[0].Feedback)
		}
	} else {
		s := make([]interface{}, len(tmp.subResults))
		for i, subRes := range tmp.subResults {
			s[i] = subRes.Value
			AddFeedback(&(pd.Result.Feedback), subRes.Feedback)
		}
		pd.Result.Value = s
	}
}

// ------- Parse multiple subparsers without lower or upper bound limit:
type ParseMulti0 struct {
	ParseMulti
}

func NewParseMulti0(parseData func(interface{}) *ParseData,
	setParseData func(interface{}, *ParseData) interface{}) *ParseMulti0 {

	p := &ParseMulti0{}
	p.parseData = parseData
	p.setParseData = setParseData
	p.ConfigPort(ParseMultiConfig{0, math.MaxInt32})
	return p
}

// ------- Parse multiple subparsers with lower bound 1 and without upper bound limit:
type ParseMulti1 struct {
	ParseMulti
}

func NewParseMulti1(parseData func(interface{}) *ParseData,
	setParseData func(interface{}, *ParseData) interface{}) *ParseMulti1 {

	p := &ParseMulti1{}
	p.parseData = parseData
	p.setParseData = setParseData
	p.ConfigPort(ParseMultiConfig{1, math.MaxInt32})
	return p
}

// ------- Parse multiple subparsers without lower bound limit and with upper bound 1:
type ParseOptional struct {
	ParseMulti
}

func NewParseOptional(parseData func(interface{}) *ParseData,
	setParseData func(interface{}, *ParseData) interface{}) *ParseOptional {

	p := &ParseOptional{}
	p.parseData = parseData
	p.setParseData = setParseData
	p.ConfigPort(ParseMultiConfig{0, 1})
	return p
}

// ------- Base for parsing multiple subparsers:
/*
ParseWithMultiSubOps is a base operation that is used in parser operations that have multiple subparsers.
The contract is the following:
We:
    - handle pd.tmp completely (so please don't touch)
    - fill tmp.pos
	and finally
	- call p.HandleSemantics

The parser op:
	- p.handleSubresult must create the matched or unmatched Result and must return nil if parsing should stop
	- p.handleSubresult *must* return a non-nil subResult if parsing should go on
	- may do some default semantics
*/
type ParseWithMultiSubOps struct {
	BaseParseOp
	subOutPorts     []func(interface{})
	handleSubresult func(interface{}, *ParseData, *tempData) *ParseResult
}

func (p *ParseWithMultiSubOps) AppendSubOutPort(subOutPort func(interface{})) {
	p.subOutPorts = append(p.subOutPorts, subOutPort)
}
func (p *ParseWithMultiSubOps) SubInPort(data interface{}) {
	pd := p.parseData(data)
	tmp := pd.tmp[len(pd.tmp)-1]
	subResult := p.handleSubresult(data, pd, tmp)
	if subResult == nil {
		pd.SubResults = tmp.subResults
		pd.tmp = pd.tmp[0 : len(pd.tmp)-1]
		p.HandleSemantics(data, pd)
	} else {
		tmp.subResults = append(tmp.subResults, subResult)
		pd.Result = nil
		p.subOutPorts[len(tmp.subResults)](data)
	}
}
func (p *ParseWithMultiSubOps) InPort(data interface{}) {
	pd := p.parseData(data)
	tmp := newTempData(pd.source.pos, len(p.subOutPorts))
	pd.tmp = append(pd.tmp, tmp)
	p.subOutPorts[0](data)
}

// ------- Parsing multiple subparsers and all have to match for a successful Result:
type ParseAll struct {
	ParseWithMultiSubOps
}

func NewParseAll(parseData func(interface{}) *ParseData,
	setParseData func(interface{}, *ParseData) interface{}) *ParseAll {

	p := &ParseAll{}
	p.parseData = parseData
	p.setParseData = setParseData
	p.handleSubresult = p.handleSubresultImpl
	return p
}
func (p *ParseAll) handleSubresultImpl(data interface{}, pd *ParseData, tmp *tempData) *ParseResult {
	switch {
	case pd.Result.ErrPos < 0 && len(tmp.subResults)+1 >= len(p.subOutPorts):
		tmp.subResults = append(tmp.subResults, pd.Result)
		pd.Result = nil
		relPos := pd.source.pos - tmp.pos
		pd.source.pos = tmp.pos
		createMatchedResult(pd, relPos)
		p.defaultSemantics(data, pd, tmp)
		return nil
	case pd.Result.ErrPos < 0:
		return pd.Result
	default:
		// pd.Result is set by subparser!
		pd.Result.Pos = tmp.pos // but we have to correct the position to the overall result position
		pd.source.pos = tmp.pos
		return nil
	}
}
func (p *ParseAll) defaultSemantics(data interface{}, pd *ParseData, tmp *tempData) {
	s := make([]interface{}, len(tmp.subResults))
	for i, subRes := range tmp.subResults {
		s[i] = subRes.Value
		AddFeedback(&(pd.Result.Feedback), subRes.Feedback)
	}
	pd.Result.Value = s
}

// ------- Parsing multiple subparsers and the first matching one delivers the successful Result:
type ParseAny struct {
	ParseWithMultiSubOps
}

func NewParseAny(parseData func(interface{}) *ParseData,
	setParseData func(interface{}, *ParseData) interface{}) *ParseAny {

	p := &ParseAny{}
	p.parseData = parseData
	p.setParseData = setParseData
	p.handleSubresult = p.handleSubresultImpl
	return p
}
func (p *ParseAny) handleSubresultImpl(data interface{}, pd *ParseData, tmp *tempData) *ParseResult {
	switch {
	case pd.Result.ErrPos < 0:
		tmp.subResults = append(tmp.subResults, pd.Result)
		pd.Result = nil
		relPos := pd.source.pos - tmp.pos
		pd.source.pos = tmp.pos
		createMatchedResult(pd, relPos)
		p.defaultSemantics(data, pd, tmp)
		return nil
	case pd.Result.ErrPos >= 0 && len(tmp.subResults)+1 >= len(p.subOutPorts):
		tmp.subResults = append(tmp.subResults, pd.Result)
		relPos := pd.Result.Pos - tmp.pos
		pd.source.pos = tmp.pos
		pd.Result = nil
		createUnmatchedResult(pd, relPos, fmt.Sprintf("Any subparser should match. But all %d subparsers failed", len(tmp.subResults)), nil)
		for _, subRes := range tmp.subResults {
			AddFeedback(&(pd.Result.Feedback), subRes.Feedback)
		}
		return nil
	default:
		pd.source.pos = tmp.pos
		return pd.Result
	}
}
func (p *ParseAny) defaultSemantics(data interface{}, pd *ParseData, tmp *tempData) {
	subRes := tmp.subResults[len(tmp.subResults)-1]
	pd.Result.Value = subRes.Value
	AddFeedback(&(pd.Result.Feedback), subRes.Feedback)
}
