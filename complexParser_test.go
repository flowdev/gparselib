package gparselib

import (
	"testing"
)

func TestParseOptional(t *testing.T) {
	pl := func(portOut func(interface{})) (portIn func(interface{})) {
		portIn = ParseLiteral(
			portOut,
			nil,
			getParseDataForTest,
			setParseDataForTest,
			"flow",
		)
		return
	}
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
	pl := func(portOut func(interface{})) (portIn func(interface{})) {
		portIn = ParseLiteral(
			portOut,
			nil,
			getParseDataForTest,
			setParseDataForTest,
			"flow",
		)
		return
	}
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
	pl := func(portOut func(interface{})) (portIn func(interface{})) {
		portIn = ParseLiteral(
			portOut,
			nil,
			getParseDataForTest,
			setParseDataForTest,
			"flow",
		)
		return
	}
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
	pl := func(portOut func(interface{})) (portIn func(interface{})) {
		portIn = ParseLiteral(
			portOut,
			nil,
			getParseDataForTest,
			setParseDataForTest,
			"flow",
		)
		return
	}
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
	pl := func(portOut func(interface{})) (portIn func(interface{})) {
		portIn = ParseLiteral(
			portOut,
			nil,
			getParseDataForTest,
			setParseDataForTest,
			"flow",
		)
		return
	}
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
