package gparselib

import (
	"reflect"
	"testing"
)

// testParseOp is the interface of all parsers to be tested.
type testParseOp func(outPort func(interface{})) (inPort func(interface{}))

func getParseDataForTest(data interface{}) *ParseData {
	return data.(*ParseData)
}

func setParseDataForTest(data interface{}, pd *ParseData) interface{} {
	return pd
}

type parseTestData struct {
	givenParseData   *ParseData
	expectedResult   *ParseResult
	expectedSrcPos   int
	expectedErrCount int
}

func TestParseLiteral(t *testing.T) {
	p := func(outPort func(interface{})) (inPort func(interface{})) {
		inPort, _, _ = ParseLiteral(
			outPort,
			nil,
			getParseDataForTest,
			setParseDataForTest,
			"func",
		)
		return
	}

	/*
		runTest(t, p, newData("no match", 0, " func\n"), newResult(0, "", nil, 0), 0, 1)
		runTest(t, p, newData("empty", 0, ""), newResult(0, "", nil, 0), 0, 1)
		runTest(t, p, newData("simple", 0, "func"), newResult(0, "func", nil, -1), 4, 0)
		runTest(t, p, newData("simple 2", 0, "func 123"), newResult(0, "func", nil, -1), 4, 0)
		runTest(t, p, newData("simple 3", 2, "12func345"), newResult(2, "func", nil, -1), 6, 0)
	*/
	runTests(t, p, []parseTestData{
		{
			givenParseData:   newData("no match", 0, " func\n"),
			expectedResult:   newResult(0, "", nil, 0),
			expectedSrcPos:   0,
			expectedErrCount: 1,
		}, {
			givenParseData:   newData("empty", 0, ""),
			expectedResult:   newResult(0, "", nil, 0),
			expectedSrcPos:   0,
			expectedErrCount: 1,
		}, {
			givenParseData:   newData("simple", 0, "func"),
			expectedResult:   newResult(0, "func", nil, -1),
			expectedSrcPos:   4,
			expectedErrCount: 0,
		}, {
			givenParseData:   newData("simple 2", 0, "func 123"),
			expectedResult:   newResult(0, "func", nil, -1),
			expectedSrcPos:   4,
			expectedErrCount: 0,
		}, {
			givenParseData:   newData("simple 3", 2, "12func345"),
			expectedResult:   newResult(2, "func", nil, -1),
			expectedSrcPos:   6,
			expectedErrCount: 0,
		},
	})
}

/*
func TestParseNatural(t *testing.T) {
	p := NewParseNatural(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd }, 10)

	runTest(t, p, newData("no match", 0, "baaa"), newResult(0, "", nil, 0), 0, 1)
	runTest(t, p, newData("empty", 0, ""), newResult(0, "", nil, 0), 0, 1)
	runTest(t, p, newData("simple", 0, "3"), newResult(0, "3", uint64(3), -1), 1, 0)
	runTest(t, p, newData("simple 2", 0, "123"), newResult(0, "123", uint64(123), -1), 3, 0)
	runTest(t, p, newData("simple 3", 2, "ab123c456"), newResult(2, "123", uint64(123), -1), 5, 0)
	runTest(t, p, newData("error", 2, "ab1234567890123456789012345678901234567890"), newResult(2, "", nil, 2), 2, 1)

	Convey("Parse natural with illegal radix, ...", t, func() {
		So(func() {
			NewParseNatural(func(data interface{}) *ParseData { return data.(*ParseData) },
				func(data interface{}, pd *ParseData) interface{} { return pd }, 1)
		}, ShouldPanic)
	})
}

func TestParseEof(t *testing.T) {
	p := NewParseEof(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd })

	runTest(t, p, newData("no match", 2, "12flow345"), newResult(2, "", nil, 2), 2, 1)
	runTest(t, p, newData("empty", 0, ""), newResult(0, "", nil, -1), 0, 0)
	runTest(t, p, newData("simple", 4, "flow"), newResult(4, "", nil, -1), 4, 0)
	runTest(t, p, newData("simple 2", 8, "flow 123"), newResult(8, "", nil, -1), 8, 0)
}

func TestParseSpace(t *testing.T) {
	pEolOk := NewParseSpace(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd }, true)
	pNotOk := NewParseSpace(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd }, false)

	runTest(t, pEolOk, newData("no match incl. EOL", 0, "ba"), newResult(0, "", nil, 0), 0, 1)
	runTest(t, pNotOk, newData("no match excl. EOL", 0, "ba"), newResult(0, "", nil, 0), 0, 1)
	runTest(t, pEolOk, newData("empty incl. EOL", 0, ""), newResult(0, "", nil, 0), 0, 1)
	runTest(t, pNotOk, newData("empty excl. EOL", 0, ""), newResult(0, "", nil, 0), 0, 1)
	runTest(t, pEolOk, newData("simple incl. EOL", 0, " "), newResult(0, " ", nil, -1), 1, 0)
	runTest(t, pNotOk, newData("simple excl. EOL", 0, " "), newResult(0, " ", nil, -1), 1, 0)
	runTest(t, pEolOk, newData("simple 2 incl. EOL", 0, " \t\r\n 123"), newResult(0, " \t\r\n ", nil, -1), 5, 0)
	runTest(t, pNotOk, newData("simple 2 excl. EOL", 0, " \t\r\n 123"), newResult(0, " \t\r", nil, -1), 3, 0)
	runTest(t, pEolOk, newData("simple 3 incl. EOL", 2, "12 \t\r\n 3456"), newResult(2, " \t\r\n ", nil, -1), 7, 0)
	runTest(t, pNotOk, newData("simple 3 excl. EOL", 2, "12 \t\r\n 3456"), newResult(2, " \t\r", nil, -1), 5, 0)
}

func TestParseRegexp(t *testing.T) {
	pWv := NewParseRegexp(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd }, `^[a]+`)
	pWoV := NewParseRegexp(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd }, `[a]+`)

	runTest(t, pWv, newData("no match with ^", 0, "baaa"), newResult(0, "", nil, 0), 0, 1)
	runTest(t, pWoV, newData("no match without ^", 0, "baaa"), newResult(0, "", nil, 0), 0, 1)
	runTest(t, pWv, newData("empty with ^", 0, ""), newResult(0, "", nil, 0), 0, 1)
	runTest(t, pWoV, newData("empty without ^", 0, ""), newResult(0, "", nil, 0), 0, 1)
	runTest(t, pWv, newData("simple with ^", 0, "a"), newResult(0, "a", "a", -1), 1, 0)
	runTest(t, pWoV, newData("simple without ^", 0, "a"), newResult(0, "a", "a", -1), 1, 0)
	runTest(t, pWv, newData("simple 2 with ^", 0, "aaa 123"), newResult(0, "aaa", "aaa", -1), 3, 0)
	runTest(t, pWoV, newData("simple 2 without ^", 0, "aaa 123"), newResult(0, "aaa", "aaa", -1), 3, 0)
	runTest(t, pWv, newData("simple 3 with ^", 2, "12aaa3456"), newResult(2, "aaa", "aaa", -1), 5, 0)
	runTest(t, pWoV, newData("simple 3 without ^", 2, "12aaa3456"), newResult(2, "aaa", "aaa", -1), 5, 0)

	Convey("Parse regexp with illegal regexp, ...", t, func() {
		So(func() {
			NewParseRegexp(func(data interface{}) *ParseData { return data.(*ParseData) },
				func(data interface{}, pd *ParseData) interface{} { return pd }, `[a`)
		}, ShouldPanic)
	})
}

func TestParseLineComment(t *testing.T) {
	p := NewParseLineComment(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd }, "//")

	runTest(t, p, newData("no match", 0, " // "), newResult(0, "", nil, 0), 0, 1)
	runTest(t, p, newData("empty", 0, ""), newResult(0, "", nil, 0), 0, 1)
	runTest(t, p, newData("simple", 0, "// 1\n"), newResult(0, "// 1", "", -1), 4, 0)
	runTest(t, p, newData("simple 2", 0, "// 1\n 23"), newResult(0, "// 1", "", -1), 4, 0)
	runTest(t, p, newData("simple 3", 2, "12// 1\n345"), newResult(2, "// 1", "", -1), 6, 0)
	runTest(t, p, newData("simple 4", 2, "12// 1\r\n345"), newResult(2, "// 1\r", "", -1), 7, 0)
	runTest(t, p, newData("evil", 0, "//"), newResult(0, "//", "", -1), 2, 0)
}
*/

//func TestParseBlockComment(t *testing.T) {
//	p := NewParseBlockComment(func(data interface{}) *ParseData { return data.(*ParseData) },
//		func(data interface{}, pd *ParseData) interface{} { return pd }, "/*", "*/")
//
//	runTest(t, p, newData("no match", 0, " 123 "), newResult(0, "", nil, 0), 0, 1)
//	runTest(t, p, newData("empty", 0, ""), newResult(0, "", nil, 0), 0, 1)
//	runTest(t, p, newData("simple", 0, "/* 123 */"), newResult(0, "/* 123 */", "", -1), 9, 0)
//	runTest(t, p, newData("simple 2", 4, "abcd/* 123 */"), newResult(4, "/* 123 */", "", -1), 13, 0)
//	runTest(t, p, newData("simple 3", 2, "ab/* 123 */cdefg"), newResult(2, "/* 123 */", "", -1), 11, 0)
//	runTest(t, p, newData("nested block comments aren't supported!!!", 2, "ab/* 1 /* 2 */ 3 */cdefg"),
//		newResult(2, "/* 1 /* 2 */", "", -1), 14, 0)
//	runTest(t, p, newData("comment not closed", 4, "abcd/* 123 "), newResult(4, "", nil, 6), 6, 1)
//	runTest(t, p, newData("comment in single qoute string", 0, `/* 1'2\'*/'3 */`),
//		newResult(0, `/* 1'2\'*/'3 */`, "", -1), 15, 0)
//	runTest(t, p, newData("comment in double qoute string", 0, `/* 1"2\"*/"3 */`),
//		newResult(0, `/* 1"2\"*/"3 */`, "", -1), 15, 0)
//	runTest(t, p, newData("comment in backqoute string", 0, "/* 1`2*/\\`*/`3 */"),
//		newResult(0, "/* 1`2*/\\`*/", "", -1), 12, 0)
//}

const semanticTestValue = "Semantic test!!!"

func TestParseSemantics(t *testing.T) {
	p := func(outPort func(interface{})) (inPort func(interface{})) {
		pInPort, pSemInPort, pSetSemOutPort := ParseLiteral(
			outPort,
			nil,
			getParseDataForTest,
			setParseDataForTest,
			"func",
		)
		semInPort := SemanticsTestOp(pSemInPort)
		pSetSemOutPort(semInPort)
		return pInPort
	}

	runTests(t, p, []parseTestData{
		{
			givenParseData:   newData("no match", 0, " func\n"),
			expectedResult:   newResult(0, "", nil, 0),
			expectedSrcPos:   0,
			expectedErrCount: 1,
		}, {
			givenParseData:   newData("simple", 0, "func"),
			expectedResult:   newResult(0, "func", semanticTestValue, -1),
			expectedSrcPos:   4,
			expectedErrCount: 0,
		},
	})
}

func newResult(pos int, text string, value interface{}, errPos int) *ParseResult {
	return &ParseResult{Pos: pos, Text: text, Value: value, ErrPos: errPos}
}
func newData(srcName string, srcPos int, srcContent string) *ParseData {
	pd := NewParseData(srcName, srcContent)
	pd.Source.pos = srcPos
	return pd
}

func runTests(t *testing.T, sp testParseOp, specs []parseTestData) {
	var pd2 *ParseData
	inPort := sp(func(data interface{}) { pd2 = data.(*ParseData) })
	for _, spec := range specs {
		t.Logf("Parsing source '%s'.", spec.givenParseData.Source.Name)
		inPort(spec.givenParseData)

		if pd2.Source.pos != spec.expectedSrcPos {
			t.Errorf("Expected source position %d, got %d.", spec.expectedSrcPos, pd2.Source.pos)
		}
		if pd2.Result == nil {
			t.Fatalf("Expected a result.")
		}
		if pd2.Result.Pos != spec.expectedResult.Pos {
			t.Errorf("Expected result position %d, got %d.", spec.expectedResult.Pos, pd2.Result.Pos)
		}
		if pd2.Result.Text != spec.expectedResult.Text {
			t.Errorf("Expected result text '%s', got '%s'.", spec.expectedResult.Text, pd2.Result.Text)
		}
		if spec.expectedResult.Value == nil && pd2.Result.Value != nil {
			t.Errorf("Didn't expect a value but got '%#v'.", pd2.Result.Value)
		}
		if spec.expectedResult.Value != nil && !reflect.DeepEqual(pd2.Result.Value, spec.expectedResult.Value) {
			t.Logf("The acutal value isn't equal to the expected one:")
			t.Errorf("Expected value of type '%T', got '%T'.", spec.expectedResult.Value, pd2.Result.Value)
			t.Errorf("Expected value '%#v', got '%#v'.", spec.expectedResult.Value, pd2.Result.Value)
		}

		if pd2.Result.ErrPos != spec.expectedResult.ErrPos {
			t.Errorf("Expected result error position %d, got %d.", spec.expectedResult.ErrPos, pd2.Result.ErrPos)
		}
		if pd2.Result.HasError() && spec.expectedErrCount <= 0 {
			t.Logf("Actual errors are: %s", printErrors(pd2.Result.Feedback))
			t.Fatalf("Expected no error but found at least one.")
		}
		if len(pd2.Result.Feedback) != spec.expectedErrCount {
			t.Logf("Actual errors are: %s", printErrors(pd2.Result.Feedback))
			t.Fatalf("Expected %d errors, got %d.", spec.expectedErrCount, len(pd2.Result.Feedback))
		}
		if spec.expectedErrCount > 0 && pd2.Result.Feedback[spec.expectedErrCount-1].Msg.String() == "" {
			t.Logf("Actual errors are: %s", printErrors(pd2.Result.Feedback))
			t.Errorf("Expected an error message.")
		}
	}
}
func printErrors(fbs []*FeedbackItem) string {
	result := ""
	for _, fb := range fbs {
		if fb.Kind == FeedbackError {
			result += fb.String() + "\n"
		}
	}
	if result == "" {
		result = "<EMPTY>"
	}
	return result
}

func SemanticsTestOp(outPort func(interface{})) (inPort func(interface{})) {
	inPort = func(data interface{}) {
		p := data.(*ParseData)
		p.Result.Value = semanticTestValue
		outPort(p)
	}
	return
}
