package gparselib

import ()

// ------- Parse multiple subparsers:

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
	p.ConfigPort(ParseMutliConfig{min, max})
	return p
}
func (p *ParseMulti) ConfigPort(cfg ParseMultiConfig) {
	p.min, p.max = cfg.Min, cfg.Max
}
func (p *ParseMulti) SetSubOutPort(subOutPort func(interface{})) {
	p.subOutPort = subOutPort
}
func (p *ParseMulti) SubInPort(data interface{}) {
}
func (p *ParseMulti) InPort(data interface{}) {
	pd := p.parseData(data)
	pos := pd.source.pos
	if len(pd.source.content) >= pos+p.cfgN && pd.source.content[pos:pos+p.cfgN] == p.cfgMulti {
		createMatchedResult(pd, p.cfgN)
	} else {
		createUnmatchedResult(pd, 0, "Literal '"+p.cfgMulti+"' expected", nil)
	}
	p.HandleSemantics(data, pd)
}
