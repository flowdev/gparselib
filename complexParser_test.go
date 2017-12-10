package gparselib

import (
	"testing"
)

func MakeParseLiteral(literal string) SubparserOp {
	return func(portOut func(interface{})) (portIn func(interface{})) {
		portIn = ParseLiteral(
			portOut,
			nil,
			getParseDataForTest,
			setParseDataForTest,
			literal,
		)
		return
	}

}

func TestParseOptional(t *testing.T) {
	pl := MakeParseLiteral("flow")
	p := func(portOut func(interface{})) (portIn func(interface{})) {
		portIn = ParseOptional(
			portOut,
			pl,
			nil,
			getParseDataForTest,
			setParseDataForTest,
		)
		return
	}

	runTests(t, p, []parseTestData{
		{
			givenParseData:   newData("no match", 0, " flow"),
			expectedResult:   newResult(0, "", nil, -1),
			expectedSrcPos:   0,
			expectedErrCount: 0,
		}, {
			givenParseData:   newData("1 match", 0, "flow"),
			expectedResult:   newResult(0, "flow", nil, -1),
			expectedSrcPos:   4,
			expectedErrCount: 0,
		}, {
			givenParseData:   newData("2 matches", 0, "flowflow"),
			expectedResult:   newResult(0, "flow", nil, -1),
			expectedSrcPos:   4,
			expectedErrCount: 0,
		},
	})
}

func TestParseMulti0(t *testing.T) {
	pl := MakeParseLiteral("flow")
	p := func(portOut func(interface{})) (portIn func(interface{})) {
		portIn = ParseMulti0(
			portOut,
			pl,
			nil,
			getParseDataForTest,
			setParseDataForTest,
		)
		return
	}

	runTests(t, p, []parseTestData{
		{
			givenParseData:   newData("no match", 0, " flow"),
			expectedResult:   newResult(0, "", []interface{}{}, -1),
			expectedSrcPos:   0,
			expectedErrCount: 0,
		}, {
			givenParseData:   newData("1 match", 0, "flow"),
			expectedResult:   newResult(0, "flow", []interface{}{nil}, -1),
			expectedSrcPos:   4,
			expectedErrCount: 0,
		}, {
			givenParseData:   newData("2 matches", 0, "flowflow"),
			expectedResult:   newResult(0, "flowflow", []interface{}{nil, nil}, -1),
			expectedSrcPos:   8,
			expectedErrCount: 0,
		},
	})
}

func TestParseMulti1(t *testing.T) {
	pl := MakeParseLiteral("flow")
	p := func(portOut func(interface{})) (portIn func(interface{})) {
		portIn = ParseMulti1(
			portOut,
			pl,
			nil,
			getParseDataForTest,
			setParseDataForTest,
		)
		return
	}

	runTests(t, p, []parseTestData{
		{
			givenParseData:   newData("no match", 0, " flow"),
			expectedResult:   newResult(0, "", nil, 0),
			expectedSrcPos:   0,
			expectedErrCount: 2,
		}, {
			givenParseData:   newData("1 match", 0, "flow"),
			expectedResult:   newResult(0, "flow", []interface{}{nil}, -1),
			expectedSrcPos:   4,
			expectedErrCount: 0,
		}, {
			givenParseData:   newData("2 matches", 0, "flowflow"),
			expectedResult:   newResult(0, "flowflow", []interface{}{nil, nil}, -1),
			expectedSrcPos:   8,
			expectedErrCount: 0,
		},
	})
}

func TestParseMulti(t *testing.T) {
	pl := MakeParseLiteral("flow")
	p := func(portOut func(interface{})) (portIn func(interface{})) {
		portIn = ParseMulti(
			portOut,
			pl,
			nil,
			getParseDataForTest,
			setParseDataForTest,
			2,
			3,
		)
		return
	}

	runTests(t, p, []parseTestData{
		{
			givenParseData:   newData("multi(2-3): 1 match", 0, "flow flow"),
			expectedResult:   newResult(0, "", nil, 4),
			expectedSrcPos:   0,
			expectedErrCount: 2,
		}, {
			givenParseData:   newData("multi(2-3): 2 matches", 0, "flowflow"),
			expectedResult:   newResult(0, "flowflow", []interface{}{nil, nil}, -1),
			expectedSrcPos:   8,
			expectedErrCount: 0,
		}, {
			givenParseData:   newData("multi(2-3): 3 matches", 0, "flowflowflow"),
			expectedResult:   newResult(0, "flowflowflow", []interface{}{nil, nil, nil}, -1),
			expectedSrcPos:   12,
			expectedErrCount: 0,
		}, {
			givenParseData:   newData("multi(2-3): 4 matches", 0, "flowflowflowflow"),
			expectedResult:   newResult(0, "flowflowflow", []interface{}{nil, nil, nil}, -1),
			expectedSrcPos:   12,
			expectedErrCount: 0,
		},
	})
}

func TestParseMultiNested(t *testing.T) {
	pl := MakeParseLiteral("flow")
	pm2to3 := func(portOut func(interface{})) (portIn func(interface{})) {
		portIn = ParseMulti(
			portOut,
			pl,
			nil,
			getParseDataForTest,
			setParseDataForTest,
			2,
			3,
		)
		return
	}
	p := func(portOut func(interface{})) (portIn func(interface{})) {
		portIn = ParseMulti1(
			portOut,
			pm2to3,
			nil,
			getParseDataForTest,
			setParseDataForTest,
		)
		return
	}

	runTests(t, p, []parseTestData{
		{
			givenParseData: newData(
				"multi(2-3) in multi(1-n): 6 match",
				0,
				"flowflowflowflowflowflowflow",
			),
			expectedResult: newResult(
				0,
				"flowflowflowflowflowflow",
				[]interface{}{
					[]interface{}{nil, nil, nil},
					[]interface{}{nil, nil, nil}},
				-1,
			),
			expectedSrcPos:   24,
			expectedErrCount: 0,
		},
	})
}

func TestParseAll_NormalFunctionality(t *testing.T) {
	plFlow := MakeParseLiteral("flow")
	plNo := MakeParseLiteral("no")
	p := func(portOut func(interface{})) (portIn func(interface{})) {
		portIn = ParseAll(
			portOut,
			[]SubparserOp{plFlow, plNo},
			nil,
			getParseDataForTest,
			setParseDataForTest,
		)
		return
	}

	runTests(t, p, []parseTestData{
		{
			givenParseData:   newData("empty", 0, ""),
			expectedResult:   newResult(0, "", nil, 0),
			expectedSrcPos:   0,
			expectedErrCount: 1,
		}, {
			givenParseData:   newData("no match", 0, " flow no"),
			expectedResult:   newResult(0, "", nil, 0),
			expectedSrcPos:   0,
			expectedErrCount: 1,
		}, {
			givenParseData:   newData("match flow", 0, "flowabc"),
			expectedResult:   newResult(0, "", nil, 4),
			expectedSrcPos:   0,
			expectedErrCount: 1,
		}, {
			givenParseData:   newData("match no", 3, "123noabc"),
			expectedResult:   newResult(3, "", nil, 3),
			expectedSrcPos:   3,
			expectedErrCount: 1,
		}, {
			givenParseData:   newData("match all", 3, "123flownoabc"),
			expectedResult:   newResult(3, "flowno", []interface{}{nil, nil}, -1),
			expectedSrcPos:   9,
			expectedErrCount: 0,
		},
	})
}

func TestParseAll_ManySubs(t *testing.T) {
	pl1 := MakeParseLiteral("1")
	pl2 := MakeParseLiteral("2")
	pl3 := MakeParseLiteral("3")
	pl4 := MakeParseLiteral("4")
	pl5 := MakeParseLiteral("5")
	pl6 := MakeParseLiteral("6")
	pl7 := MakeParseLiteral("7")
	pl8 := MakeParseLiteral("8")
	pl9 := MakeParseLiteral("9")
	p := func(portOut func(interface{})) (portIn func(interface{})) {
		portIn = ParseAll(
			portOut,
			[]SubparserOp{pl1, pl2, pl3, pl4, pl5, pl6, pl7, pl8, pl9},
			nil,
			getParseDataForTest,
			setParseDataForTest,
		)
		return
	}

	runTests(t, p, []parseTestData{
		{
			givenParseData: newData("match 9", 0, "1234567890"),
			expectedResult: newResult(
				0,
				"123456789",
				[]interface{}{nil, nil, nil, nil, nil, nil, nil, nil, nil},
				-1,
			),
			expectedSrcPos:   9,
			expectedErrCount: 0,
		},
	})
}

func TestParseAll_Nested(t *testing.T) {
	plFlow := MakeParseLiteral("flow")
	plNo := MakeParseLiteral("no")
	pInner := func(portOut func(interface{})) (portIn func(interface{})) {
		portIn = ParseAll(
			portOut,
			[]SubparserOp{plFlow, plNo},
			nil,
			getParseDataForTest,
			setParseDataForTest,
		)
		return
	}
	plFun := MakeParseLiteral("fun")
	pOuter := func(portOut func(interface{})) (portIn func(interface{})) {
		portIn = ParseAll(
			portOut,
			[]SubparserOp{plFun, pInner},
			nil,
			getParseDataForTest,
			setParseDataForTest,
		)
		return
	}

	runTests(t, pOuter, []parseTestData{
		{
			givenParseData: newData("match nested", 3, "123funflownoabc"),
			expectedResult: newResult(
				3,
				"funflowno",
				[]interface{}{nil, []interface{}{nil, nil}},
				-1,
			),
			expectedSrcPos:   12,
			expectedErrCount: 0,
		},
	})
}

func TestParseAny_NormalFunctionality(t *testing.T) {
	plFlow := MakeParseLiteral("flow")
	plNo := MakeParseLiteral("no")
	p := func(portOut func(interface{})) (portIn func(interface{})) {
		portIn = ParseAny(
			portOut,
			[]SubparserOp{plFlow, plNo},
			nil,
			getParseDataForTest,
			setParseDataForTest,
		)
		return
	}

	runTests(t, p, []parseTestData{
		{
			givenParseData:   newData("empty", 0, ""),
			expectedResult:   newResult(0, "", nil, 0),
			expectedSrcPos:   0,
			expectedErrCount: 3,
		}, {
			givenParseData:   newData("no match", 0, " flow 3"),
			expectedResult:   newResult(0, "", nil, 0),
			expectedSrcPos:   0,
			expectedErrCount: 3,
		}, {
			givenParseData:   newData("match flow", 0, "flowabc"),
			expectedResult:   newResult(0, "flow", nil, -1),
			expectedSrcPos:   4,
			expectedErrCount: 0,
		}, {
			givenParseData:   newData("match no", 2, "12noabc"),
			expectedResult:   newResult(2, "no", nil, -1),
			expectedSrcPos:   4,
			expectedErrCount: 0,
		}, {
			givenParseData:   newData("match both", 3, "123flownoabc"),
			expectedResult:   newResult(3, "flow", nil, -1),
			expectedSrcPos:   7,
			expectedErrCount: 0,
		},
	})
}

func TestParseAny_Nested(t *testing.T) {
	plFlow := MakeParseLiteral("flow")
	plNo := MakeParseLiteral("no")
	pInner := func(portOut func(interface{})) (portIn func(interface{})) {
		portIn = ParseAny(
			portOut,
			[]SubparserOp{plFlow, plNo},
			nil,
			getParseDataForTest,
			setParseDataForTest,
		)
		return
	}

	plFun := MakeParseLiteral("fun")
	pOuter := func(portOut func(interface{})) (portIn func(interface{})) {
		portIn = ParseAny(
			portOut,
			[]SubparserOp{plFun, pInner},
			nil,
			getParseDataForTest,
			setParseDataForTest,
		)
		return
	}

	runTests(t, pOuter, []parseTestData{
		{
			givenParseData:   newData("match flow", 3, "123flowabc"),
			expectedResult:   newResult(3, "flow", nil, -1),
			expectedSrcPos:   7,
			expectedErrCount: 0,
		}, {
			givenParseData:   newData("match fun", 3, "123funabc"),
			expectedResult:   newResult(3, "fun", nil, -1),
			expectedSrcPos:   6,
			expectedErrCount: 0,
		}, {
			givenParseData:   newData("match no", 3, "123noabc"),
			expectedResult:   newResult(3, "no", nil, -1),
			expectedSrcPos:   5,
			expectedErrCount: 0,
		},
	})
}
