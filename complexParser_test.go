package gparselib

import (
	"testing"
)

func TestParseMulti(t *testing.T) {
	pl := NewParseLiteral(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd }, "flow")
	p0_1 := NewParseMulti(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd }, 0, 1)
	p0_1.SetSemOutPort(pl.InPort)
	pl.SetOutPort(p0_1.SemInPort)
	p0_n := NewParseMulti(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd }, 0, 1)
	p1_n := NewParseMulti(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd }, 1, 1)
	p2_3 := NewParseMulti(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd }, 2, 3)

	runTest(t, p0_1, newData("0-1: no match", 0, " flow"), newResult(0, "", nil, -1), 0, 0)
	runTest(t, p0_1, newData("0-1: 1 match", 0, "flow"), newResult(0, "flow", nil, -1), 4, 0)
	runTest(t, p0_1, newData("0-1: 2 match", 0, "flowflow"), newResult(0, "flow", nil, -1), 4, 0)
	runTest(t, p0_n, newData("0-n: no match", 0, " flow"), newResult(0, "", nil, -1), 0, 0)
	runTest(t, p0_n, newData("0-n: 1 match", 0, "flow"), newResult(0, "flow", nil, -1), 4, 0)
	runTest(t, p0_n, newData("0-n: 2 match", 0, "flowflow"), newResult(0, "flowflow", nil, -1), 8, 0)
	runTest(t, p1_n, newData("1-n: no match", 0, " flow"), newResult(0, "", nil, 0), 0, 1)
	runTest(t, p1_n, newData("1-n: 1 match", 0, "flow"), newResult(0, "flow", nil, -1), 4, 0)
	runTest(t, p1_n, newData("1-n: 2 match", 0, "flowflow"), newResult(0, "flowflow", nil, -1), 8, 0)
	runTest(t, p1_n, newData("1-n: 5 match", 0, "flowflowflowflowflow"), newResult(0, "flowflowflowflowflow", nil, -1), 20, 0)
	runTest(t, p2_3, newData("2-3: 1 match", 0, "flow flow"), newResult(0, "", nil, 4), 0, 1)
	runTest(t, p2_3, newData("2-3: 2 match", 0, "flowflow"), newResult(0, "flowflow", nil, -1), 8, 0)
	runTest(t, p2_3, newData("2-3: 3 match", 0, "flowflowflow"), newResult(0, "flowflowflow", nil, -1), 12, 0)
	runTest(t, p2_3, newData("2-3: 4 match", 0, "flowflowflowflow"), newResult(0, "flowflowflow", nil, -1), 12, 0)
}
