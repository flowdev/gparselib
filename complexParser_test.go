package gparselib

import (
	"math"
	"testing"
)

func TestParseMulti(t *testing.T) {
	pl := NewParseLiteral(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd }, "flow")
	p0_1 := NewParseMulti(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd }, 0, 1)
	p0_1.SetSubOutPort(pl.InPort)
	pl.SetOutPort(p0_1.SubInPort)
	p0_n := NewParseMulti(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd }, 0, math.MaxInt32)
	p1_n := NewParseMulti(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd }, 1, math.MaxInt32)
	p2_3 := NewParseMulti(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd }, 2, 3)

	runTest(t, p0_1, newData("0-1: no match", 0, " flow"), newResult(0, "", nil, -1), 0, 0)
	runTest(t, p0_1, newData("0-1: 1 match", 0, "flow"), newResult(0, "flow", nil, -1), 4, 0)
	runTest(t, p0_1, newData("0-1: 2 match", 0, "flowflow"), newResult(0, "flow", nil, -1), 4, 0)
	p0_n.SetSubOutPort(pl.InPort)
	pl.SetOutPort(p0_n.SubInPort)
	runTest(t, p0_n, newData("0-n: no match", 0, " flow"), newResult(0, "", []interface{}{}, -1), 0, 0)
	runTest(t, p0_n, newData("0-n: 1 match", 0, "flow"), newResult(0, "flow", []interface{}{nil}, -1), 4, 0)
	runTest(t, p0_n, newData("0-n: 2 match", 0, "flowflow"),
		newResult(0, "flowflow", []interface{}{nil, nil}, -1), 8, 0)
	p1_n.SetSubOutPort(pl.InPort)
	pl.SetOutPort(p1_n.SubInPort)
	runTest(t, p1_n, newData("1-n: no match", 0, " flow"), newResult(0, "", nil, 0), 0, 2)
	runTest(t, p1_n, newData("1-n: 1 match", 0, "flow"), newResult(0, "flow", []interface{}{nil}, -1), 4, 0)
	runTest(t, p1_n, newData("1-n: 2 match", 0, "flowflow"),
		newResult(0, "flowflow", []interface{}{nil, nil}, -1), 8, 0)
	runTest(t, p1_n, newData("1-n: 5 match", 0, "flowflowflowflowflow"),
		newResult(0, "flowflowflowflowflow", []interface{}{nil, nil, nil, nil, nil}, -1), 20, 0)
	p2_3.SetSubOutPort(pl.InPort)
	pl.SetOutPort(p2_3.SubInPort)
	runTest(t, p2_3, newData("2-3: 1 match", 0, "flow flow"), newResult(0, "", nil, 4), 0, 2)
	runTest(t, p2_3, newData("2-3: 2 match", 0, "flowflow"),
		newResult(0, "flowflow", []interface{}{nil, nil}, -1), 8, 0)
	runTest(t, p2_3, newData("2-3: 3 match", 0, "flowflowflow"),
		newResult(0, "flowflowflow", []interface{}{nil, nil, nil}, -1), 12, 0)
	runTest(t, p2_3, newData("2-3: 4 match", 0, "flowflowflowflow"),
		newResult(0, "flowflowflow", []interface{}{nil, nil, nil}, -1), 12, 0)
	p1_n.SetSubOutPort(p2_3.InPort)
	p2_3.SetOutPort(p1_n.SubInPort)
	runTest(t, p1_n, newData("2-3 in 1-n: 6 match", 0, "flowflowflowflowflowflowflow"),
		newResult(0, "flowflowflowflowflowflow",
			[]interface{}{[]interface{}{nil, nil, nil}, []interface{}{nil, nil, nil}}, -1), 24, 0)
}

func TestParseMulti0(t *testing.T) {
	pl := NewParseLiteral(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd }, "flow")
	p := NewParseMulti0(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd })
	p.SetSubOutPort(pl.InPort)
	pl.SetOutPort(p.SubInPort)

	runTest(t, p, newData("no match", 0, " flow"), newResult(0, "", []interface{}{}, -1), 0, 0)
	runTest(t, p, newData("1 match", 0, "flow"), newResult(0, "flow", []interface{}{nil}, -1), 4, 0)
	runTest(t, p, newData("2 match", 0, "flowflow"), newResult(0, "flowflow", []interface{}{nil, nil}, -1), 8, 0)
}

func TestParseMulti1(t *testing.T) {
	pl := NewParseLiteral(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd }, "flow")
	p := NewParseMulti1(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd })
	p.SetSubOutPort(pl.InPort)
	pl.SetOutPort(p.SubInPort)

	runTest(t, p, newData("no match", 0, " flow"), newResult(0, "", nil, 0), 0, 2)
	runTest(t, p, newData("1 match", 0, "flow"), newResult(0, "flow", []interface{}{nil}, -1), 4, 0)
	runTest(t, p, newData("2 match", 0, "flowflow"), newResult(0, "flowflow", []interface{}{nil, nil}, -1), 8, 0)
}

func TestParseOptional(t *testing.T) {
	pl := NewParseLiteral(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd }, "flow")
	p := NewParseOptional(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd })
	p.SetSubOutPort(pl.InPort)
	pl.SetOutPort(p.SubInPort)

	runTest(t, p, newData("no match", 0, " flow"), newResult(0, "", nil, -1), 0, 0)
	runTest(t, p, newData("1 match", 0, "flow"), newResult(0, "flow", nil, -1), 4, 0)
	runTest(t, p, newData("2 match", 0, "flowflow"), newResult(0, "flow", nil, -1), 4, 0)
}

func TestParseAll_NormalFunctionality(t *testing.T) {
	plFlow := NewParseLiteral(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd }, "flow")
	plNo := NewParseLiteral(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd }, "no")
	p := NewParseAll(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd })
	p.AppendSubOutPort(plFlow.InPort)
	plFlow.SetOutPort(p.SubInPort)
	p.AppendSubOutPort(plNo.InPort)
	plNo.SetOutPort(p.SubInPort)

	runTest(t, p, newData("empty", 0, ""), newResult(0, "", nil, 0), 0, 2)
	runTest(t, p, newData("no match", 0, " flow no"), newResult(0, "", nil, 0), 0, 2)
	runTest(t, p, newData("match flow", 0, "flowabc"), newResult(0, "", nil, 4), 0, 2)
	runTest(t, p, newData("match no", 3, "123noabc"), newResult(3, "", nil, 3), 3, 2)
	runTest(t, p, newData("match all", 3, "123flownoabc"), newResult(3, "flowno", []interface{}{nil, nil}, -1), 9, 0)
}
func TestParseAll_ManySubs(t *testing.T) {
	pl1 := NewParseLiteral(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd }, "1")
	pl2 := NewParseLiteral(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd }, "2")
	pl3 := NewParseLiteral(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd }, "3")
	pl4 := NewParseLiteral(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd }, "4")
	pl5 := NewParseLiteral(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd }, "5")
	pl6 := NewParseLiteral(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd }, "6")
	pl7 := NewParseLiteral(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd }, "7")
	pl8 := NewParseLiteral(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd }, "8")
	pl9 := NewParseLiteral(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd }, "9")
	p := NewParseAll(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd })

	p.AppendSubOutPort(pl1.InPort)
	pl1.SetOutPort(p.SubInPort)
	p.AppendSubOutPort(pl2.InPort)
	pl2.SetOutPort(p.SubInPort)
	p.AppendSubOutPort(pl3.InPort)
	pl3.SetOutPort(p.SubInPort)
	p.AppendSubOutPort(pl4.InPort)
	pl4.SetOutPort(p.SubInPort)
	p.AppendSubOutPort(pl5.InPort)
	pl5.SetOutPort(p.SubInPort)
	p.AppendSubOutPort(pl6.InPort)
	pl6.SetOutPort(p.SubInPort)
	p.AppendSubOutPort(pl7.InPort)
	pl7.SetOutPort(p.SubInPort)
	p.AppendSubOutPort(pl8.InPort)
	pl8.SetOutPort(p.SubInPort)
	p.AppendSubOutPort(pl9.InPort)
	pl9.SetOutPort(p.SubInPort)

	runTest(t, p, newData("match 9", 0, "1234567890"),
		newResult(0, "123456789", []interface{}{nil, nil, nil, nil, nil, nil, nil, nil, nil}, -1), 9, 0)
}
func TestParseAll_Nested(t *testing.T) {
	plFlow := NewParseLiteral(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd }, "flow")
	plNo := NewParseLiteral(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd }, "no")
	pInner := NewParseAll(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd })
	plFun := NewParseLiteral(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd }, "fun")
	pOuter := NewParseAll(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd })
	pInner.AppendSubOutPort(plFlow.InPort)
	plFlow.SetOutPort(pInner.SubInPort)
	pInner.AppendSubOutPort(plNo.InPort)
	plNo.SetOutPort(pInner.SubInPort)
	pOuter.AppendSubOutPort(plFun.InPort)
	plFun.SetOutPort(pOuter.SubInPort)
	pOuter.AppendSubOutPort(pInner.InPort)
	pInner.SetOutPort(pOuter.SubInPort)

	runTest(t, pOuter, newData("match nested", 3, "123funflownoabc"),
		newResult(3, "funflowno", []interface{}{nil, []interface{}{nil, nil}}, -1), 12, 0)
}

func TestParseAny_NormalFunctionality(t *testing.T) {
	plFlow := NewParseLiteral(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd }, "flow")
	pn := NewParseNatural(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd }, 10)
	p := NewParseAny(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd })
	p.AppendSubOutPort(plFlow.InPort)
	plFlow.SetOutPort(p.SubInPort)
	p.AppendSubOutPort(pn.InPort)
	pn.SetOutPort(p.SubInPort)

	runTest(t, p, newData("empty", 0, ""), newResult(0, "", nil, 0), 0, 3)
	runTest(t, p, newData("no match", 0, " flow 3"), newResult(0, "", nil, 0), 0, 3)
	runTest(t, p, newData("match flow", 0, "flowabc"), newResult(0, "flow", nil, -1), 4, 0)
	runTest(t, p, newData("match 3", 2, "123noabc"), newResult(2, "3", uint64(3), -1), 3, 0)
	runTest(t, p, newData("match both", 3, "123flow3abc"), newResult(3, "flow", nil, -1), 7, 0)
}
func TestParseAny_Nested(t *testing.T) {
	plFlow := NewParseLiteral(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd }, "flow")
	pn := NewParseNatural(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd }, 10)
	pInner := NewParseAny(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd })
	pInner.AppendSubOutPort(plFlow.InPort)
	plFlow.SetOutPort(pInner.SubInPort)
	pInner.AppendSubOutPort(pn.InPort)
	pn.SetOutPort(pInner.SubInPort)
	plFun := NewParseLiteral(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd }, "fun")
	pOuter := NewParseAny(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd })
	pOuter.AppendSubOutPort(plFun.InPort)
	plFun.SetOutPort(pOuter.SubInPort)
	pOuter.AppendSubOutPort(pInner.InPort)
	pInner.SetOutPort(pOuter.SubInPort)

	runTest(t, pOuter, newData("match flow", 3, "123flowabc"), newResult(3, "flow", nil, -1), 7, 0)
	runTest(t, pOuter, newData("match fun", 3, "123funabc"), newResult(3, "fun", nil, -1), 6, 0)
	runTest(t, pOuter, newData("match 3", 3, "1233abc"), newResult(3, "3", uint64(3), -1), 4, 0)
}
