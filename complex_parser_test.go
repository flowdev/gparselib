package gparselib

import (
	"testing"
)

func MakeParseLiteral(literal string) SubparserOp {
	return func(pd *ParseData, ctx interface{}) (*ParseData, interface{}) {
		return ParseLiteral(pd, ctx, nil, literal)
	}
}

func TestParseOptional(t *testing.T) {
	p := func(pd *ParseData, ctx interface{}) (*ParseData, interface{}) {
		return ParseOptional(pd, ctx, MakeParseLiteral("flow"), nil)
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
	p := func(pd *ParseData, ctx interface{}) (*ParseData, interface{}) {
		return ParseMulti0(pd, ctx, MakeParseLiteral("flow"), nil)
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
	p := func(pd *ParseData, ctx interface{}) (*ParseData, interface{}) {
		return ParseMulti1(pd, ctx, MakeParseLiteral("flow"), nil)
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
	p := func(pd *ParseData, ctx interface{}) (*ParseData, interface{}) {
		return ParseMulti(pd, ctx, MakeParseLiteral("flow"), nil, 2, 3)
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
	pm2to3 := func(pd *ParseData, ctx interface{}) (*ParseData, interface{}) {
		return ParseMulti(pd, ctx, MakeParseLiteral("flow"), nil, 2, 3)
	}
	p := func(pd *ParseData, ctx interface{}) (*ParseData, interface{}) {
		return ParseMulti1(pd, ctx, pm2to3, nil)
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
	p := func(pd *ParseData, ctx interface{}) (*ParseData, interface{}) {
		return ParseAll(pd, ctx, []SubparserOp{plFlow, plNo}, nil)
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
	p := func(pd *ParseData, ctx interface{}) (*ParseData, interface{}) {
		return ParseAll(pd, ctx,
			[]SubparserOp{pl1, pl2, pl3, pl4, pl5, pl6, pl7, pl8, pl9}, nil)
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
	pInner := func(pd *ParseData, ctx interface{}) (*ParseData, interface{}) {
		return ParseAll(pd, ctx, []SubparserOp{plFlow, plNo}, nil)
	}
	plFun := MakeParseLiteral("fun")
	pOuter := func(pd *ParseData, ctx interface{}) (*ParseData, interface{}) {
		return ParseAll(pd, ctx, []SubparserOp{plFun, pInner}, nil)
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
	p := func(pd *ParseData, ctx interface{}) (*ParseData, interface{}) {
		return ParseAny(pd, ctx, []SubparserOp{plFlow, plNo}, nil)
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
	pInner := func(pd *ParseData, ctx interface{}) (*ParseData, interface{}) {
		return ParseAny(pd, ctx, []SubparserOp{plFlow, plNo}, nil)
	}

	plFun := MakeParseLiteral("fun")
	pOuter := func(pd *ParseData, ctx interface{}) (*ParseData, interface{}) {
		return ParseAny(pd, ctx, []SubparserOp{plFun, pInner}, nil)
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
