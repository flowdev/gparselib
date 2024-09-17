package gparselib

import (
	"reflect"
	"testing"
)

type parseTestData struct {
	givenParseData   *ParseData
	expectedResult   *ParseResult
	expectedSrcPos   int
	expectedErrCount int
}

func TestParseLiteral(t *testing.T) {
	p := NewParseLiteralPlugin(nil, "func")

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

func TestParseIdent(t *testing.T) {
	p := NewParseIdentPlugin(nil, "_", "_-")

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
			givenParseData:   newData("simple", 0, "myFunc"),
			expectedResult:   newResult(0, "myFunc", nil, -1),
			expectedSrcPos:   6,
			expectedErrCount: 0,
		}, {
			givenParseData:   newData("simple 2", 0, "bla231 123"),
			expectedResult:   newResult(0, "bla231", nil, -1),
			expectedSrcPos:   6,
			expectedErrCount: 0,
		}, {
			givenParseData:   newData("hyphen+underscore", 2, "12_fu-nc3+45"),
			expectedResult:   newResult(2, "_fu-nc3", nil, -1),
			expectedSrcPos:   9,
			expectedErrCount: 0,
		}, {
			givenParseData:   newData("all allowed", 2, "12_f_u-n3cöt+45"),
			expectedResult:   newResult(2, "_f_u-n3cöt", nil, -1),
			expectedSrcPos:   13,
			expectedErrCount: 0,
		}, {
			givenParseData:   newData("hyphen at start", 2, "12-nc3+45"),
			expectedResult:   newResult(2, "", nil, 2),
			expectedSrcPos:   2,
			expectedErrCount: 1,
		},
	})
}

func TestParseNatural(t *testing.T) {
	p, _ := NewParseNaturalPlugin(nil, 10)

	runTests(t, p, []parseTestData{
		{
			givenParseData:   newData("empty", 0, ""),
			expectedResult:   newResult(0, "", nil, 0),
			expectedSrcPos:   0,
			expectedErrCount: 1,
		}, {
			givenParseData:   newData("simple", 0, "3"),
			expectedResult:   newResult(0, "3", uint64(3), -1),
			expectedSrcPos:   1,
			expectedErrCount: 0,
		}, {
			givenParseData:   newData("simple 2", 0, "123"),
			expectedResult:   newResult(0, "123", uint64(123), -1),
			expectedSrcPos:   3,
			expectedErrCount: 0,
		}, {
			givenParseData:   newData("simple 3", 2, "ab123c456"),
			expectedResult:   newResult(2, "123", uint64(123), -1),
			expectedSrcPos:   5,
			expectedErrCount: 0,
		}, {
			givenParseData: newData(
				"error",
				2,
				"ab1234567890123456789012345678901234567890",
			),
			expectedResult:   newResult(2, "", nil, 2),
			expectedSrcPos:   2,
			expectedErrCount: 1,
		},
	})

	_, err := NewParseNaturalPlugin(nil, 37)
	if err == nil || err.Error() == "" {
		t.Errorf("Expected an error with a message.")
	}
}

func TestParseEOF(t *testing.T) {
	p := NewParseEOFPlugin(nil)

	runTests(t, p, []parseTestData{
		{
			givenParseData:   newData("no match", 2, "12flow345"),
			expectedResult:   newResult(2, "", nil, 2),
			expectedSrcPos:   2,
			expectedErrCount: 1,
		}, {
			givenParseData:   newData("no match 2", 4, "flow1"),
			expectedResult:   newResult(4, "", nil, 4),
			expectedSrcPos:   4,
			expectedErrCount: 1,
		}, {
			givenParseData:   newData("empty", 0, ""),
			expectedResult:   newResult(0, "", nil, -1),
			expectedSrcPos:   0,
			expectedErrCount: 0,
		}, {
			givenParseData:   newData("simple", 4, "flow"),
			expectedResult:   newResult(4, "", nil, -1),
			expectedSrcPos:   4,
			expectedErrCount: 0,
		}, {
			givenParseData:   newData("simple 2", 8, "flow 123"),
			expectedResult:   newResult(8, "", nil, -1),
			expectedSrcPos:   8,
			expectedErrCount: 0,
		},
	})
}

func TestParseSpace(t *testing.T) {
	pEOLOK := NewParseSpacePlugin(nil, true)
	pEOLNotOK := NewParseSpacePlugin(nil, false)

	runTests(t, pEOLOK, []parseTestData{
		{
			givenParseData:   newData("no match incl. EOL", 0, "ba"),
			expectedResult:   newResult(0, "", nil, 0),
			expectedSrcPos:   0,
			expectedErrCount: 1,
		}, {
			givenParseData:   newData("empty incl. EOL", 0, ""),
			expectedResult:   newResult(0, "", nil, 0),
			expectedSrcPos:   0,
			expectedErrCount: 1,
		}, {
			givenParseData:   newData("simple incl. EOL", 0, " "),
			expectedResult:   newResult(0, " ", nil, -1),
			expectedSrcPos:   1,
			expectedErrCount: 0,
		}, {
			givenParseData:   newData("simple 2 incl. EOL", 0, " \t\r\n 123"),
			expectedResult:   newResult(0, " \t\r\n ", nil, -1),
			expectedSrcPos:   5,
			expectedErrCount: 0,
		}, {
			givenParseData: newData(
				"simple 3 incl. EOL",
				2,
				"12 \t\r\n 3456",
			),
			expectedResult:   newResult(2, " \t\r\n ", nil, -1),
			expectedSrcPos:   7,
			expectedErrCount: 0,
		},
	})
	runTests(t, pEOLNotOK, []parseTestData{
		{
			givenParseData:   newData("no match excl. EOL", 0, "ba"),
			expectedResult:   newResult(0, "", nil, 0),
			expectedSrcPos:   0,
			expectedErrCount: 1,
		}, {
			givenParseData:   newData("empty excl. EOL", 0, ""),
			expectedResult:   newResult(0, "", nil, 0),
			expectedSrcPos:   0,
			expectedErrCount: 1,
		}, {
			givenParseData:   newData("simple excl. EOL", 0, " "),
			expectedResult:   newResult(0, " ", nil, -1),
			expectedSrcPos:   1,
			expectedErrCount: 0,
		}, {
			givenParseData:   newData("simple 2 excl. EOL", 0, " \t\r\n 123"),
			expectedResult:   newResult(0, " \t\r", nil, -1),
			expectedSrcPos:   3,
			expectedErrCount: 0,
		}, {
			givenParseData: newData(
				"simple 3 excl. EOL",
				2,
				"12 \t\r\n 3456",
			),
			expectedResult:   newResult(2, " \t\r", nil, -1),
			expectedSrcPos:   5,
			expectedErrCount: 0,
		},
	})
}

func TestParseRegexp(t *testing.T) {
	pWiV, _ := NewParseRegexpPlugin(nil, `^[a]+`)
	pWoV, _ := NewParseRegexpPlugin(nil, `[a]+`)

	runTests(t, pWiV, []parseTestData{
		{
			givenParseData:   newData("no match with ^", 0, "baaa"),
			expectedResult:   newResult(0, "", nil, 0),
			expectedSrcPos:   0,
			expectedErrCount: 1,
		}, {
			givenParseData:   newData("empty with ^", 0, ""),
			expectedResult:   newResult(0, "", nil, 0),
			expectedSrcPos:   0,
			expectedErrCount: 1,
		}, {
			givenParseData:   newData("simple with ^", 0, "a"),
			expectedResult:   newResult(0, "a", "a", -1),
			expectedSrcPos:   1,
			expectedErrCount: 0,
		}, {
			givenParseData:   newData("simple 2 with ^", 0, "aaa 123"),
			expectedResult:   newResult(0, "aaa", "aaa", -1),
			expectedSrcPos:   3,
			expectedErrCount: 0,
		}, {
			givenParseData:   newData("simple 3 with ^", 2, "12aaa3456"),
			expectedResult:   newResult(2, "aaa", "aaa", -1),
			expectedSrcPos:   5,
			expectedErrCount: 0,
		},
	})
	runTests(t, pWoV, []parseTestData{
		{
			givenParseData:   newData("no match without ^", 0, "baaa"),
			expectedResult:   newResult(0, "", nil, 0),
			expectedSrcPos:   0,
			expectedErrCount: 1,
		}, {
			givenParseData:   newData("empty without ^", 0, ""),
			expectedResult:   newResult(0, "", nil, 0),
			expectedSrcPos:   0,
			expectedErrCount: 1,
		}, {
			givenParseData:   newData("simple without ^", 0, "a"),
			expectedResult:   newResult(0, "a", "a", -1),
			expectedSrcPos:   1,
			expectedErrCount: 0,
		}, {
			givenParseData:   newData("simple 2 without ^", 0, "aaa 123"),
			expectedResult:   newResult(0, "aaa", "aaa", -1),
			expectedSrcPos:   3,
			expectedErrCount: 0,
		}, {
			givenParseData:   newData("simple 3 without ^", 2, "12aaa3456"),
			expectedResult:   newResult(2, "aaa", "aaa", -1),
			expectedSrcPos:   5,
			expectedErrCount: 0,
		},
	})

	_, err := NewParseRegexpPlugin(nil, `[a`)
	if err == nil || err.Error() == "" {
		t.Errorf("Expected an error with a message.")
	}
}

func TestParseLineComment(t *testing.T) {
	p, _ := NewParseLineCommentPlugin(nil, `//`)

	runTests(t, p, []parseTestData{
		{
			givenParseData:   newData("no match", 0, " // "),
			expectedResult:   newResult(0, "", nil, 0),
			expectedSrcPos:   0,
			expectedErrCount: 1,
		}, {
			givenParseData:   newData("empty", 0, ""),
			expectedResult:   newResult(0, "", nil, 0),
			expectedSrcPos:   0,
			expectedErrCount: 1,
		}, {
			givenParseData:   newData("simple", 0, "// 1\n"),
			expectedResult:   newResult(0, "// 1", "", -1),
			expectedSrcPos:   4,
			expectedErrCount: 0,
		}, {
			givenParseData:   newData("simple 2", 0, "// 1\n 23"),
			expectedResult:   newResult(0, "// 1", "", -1),
			expectedSrcPos:   4,
			expectedErrCount: 0,
		}, {
			givenParseData:   newData("simple 3", 2, "12// 1\n345"),
			expectedResult:   newResult(2, "// 1", "", -1),
			expectedSrcPos:   6,
			expectedErrCount: 0,
		}, {
			givenParseData:   newData("simple 4", 2, "12// 1\r\n345"),
			expectedResult:   newResult(2, "// 1\r", "", -1),
			expectedSrcPos:   7,
			expectedErrCount: 0,
		}, {
			givenParseData:   newData("evil", 0, "//"),
			expectedResult:   newResult(0, "//", "", -1),
			expectedSrcPos:   2,
			expectedErrCount: 0,
		},
	})

	_, err := NewParseLineCommentPlugin(nil, ``)
	if err == nil || err.Error() == "" {
		t.Errorf("Expected an error with a message.")
	}
}

func TestParseBlockComment(t *testing.T) {
	p, _ := NewParseBlockCommentPlugin(nil, `/*`, `*/`)

	runTests(t, p, []parseTestData{
		{
			givenParseData:   newData("no match", 0, " 123 "),
			expectedResult:   newResult(0, "", nil, 0),
			expectedSrcPos:   0,
			expectedErrCount: 1,
		}, {
			givenParseData:   newData("empty", 0, ""),
			expectedResult:   newResult(0, "", nil, 0),
			expectedSrcPos:   0,
			expectedErrCount: 1,
		}, {
			givenParseData:   newData("simple", 0, "/* 123 */"),
			expectedResult:   newResult(0, "/* 123 */", "", -1),
			expectedSrcPos:   9,
			expectedErrCount: 0,
		}, {
			givenParseData:   newData("simple 2", 4, "abcd/* 123 */"),
			expectedResult:   newResult(4, "/* 123 */", "", -1),
			expectedSrcPos:   13,
			expectedErrCount: 0,
		}, {
			givenParseData:   newData("simple 3", 2, "ab/* 123 */cdefg"),
			expectedResult:   newResult(2, "/* 123 */", "", -1),
			expectedSrcPos:   11,
			expectedErrCount: 0,
		}, {
			givenParseData: newData(
				"nested block comments aren't supported!!!",
				2,
				"ab/* 1 /* 2 */ 3 */cdefg",
			),
			expectedResult:   newResult(2, "/* 1 /* 2 */", "", -1),
			expectedSrcPos:   14,
			expectedErrCount: 0,
		}, {
			givenParseData:   newData("comment not closed", 4, "abcd/* 123 "),
			expectedResult:   newResult(4, "", nil, 6),
			expectedSrcPos:   6,
			expectedErrCount: 1,
		}, {
			givenParseData: newData(
				"comment in single qoute string",
				0,
				`/* 1'2\'*/'3 */`,
			),
			expectedResult:   newResult(0, `/* 1'2\'*/'3 */`, "", -1),
			expectedSrcPos:   15,
			expectedErrCount: 0,
		}, {
			givenParseData: newData(
				"comment in double qoute string",
				0,
				`/* 1"2\"*/"3 */`,
			),
			expectedResult:   newResult(0, `/* 1"2\"*/"3 */`, "", -1),
			expectedSrcPos:   15,
			expectedErrCount: 0,
		}, {
			givenParseData: newData(
				"comment in backqoute string",
				0,
				"/* 1`2*/\\`*/`3 */",
			),
			expectedResult:   newResult(0, "/* 1`2*/\\`*/", "", -1),
			expectedSrcPos:   12,
			expectedErrCount: 0,
		},
	})

	_, err := NewParseBlockCommentPlugin(nil, ``, `*/`)
	if err == nil || err.Error() == "" {
		t.Errorf("Expected an error with a message for missing comment start.")
	}
	_, err = NewParseBlockCommentPlugin(nil, `/*`, ``)
	if err == nil || err.Error() == "" {
		t.Errorf("Expected an error with a message for missing comment end.")
	}
}

const semanticTestValue = "Semantic test!!!"

func TestParseSemantics(t *testing.T) {
	p := NewParseLiteralPlugin(SemanticsTestOp, "func")

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

func newResult(
	pos int,
	text string,
	value interface{},
	errPos int,
) *ParseResult {
	return &ParseResult{Pos: pos, Text: text, Value: value, ErrPos: errPos}
}

func newData(srcName string, srcPos int, srcContent string) *ParseData {
	pd := NewParseData(srcName, srcContent)
	pd.Source.pos = srcPos
	return pd
}

func runTests(t *testing.T, sp SubparserOp, specs []parseTestData) {
	var pd2 *ParseData
	for _, spec := range specs {
		t.Logf("Parsing source '%s'.", spec.givenParseData.Source.Name)
		pd2, _ = sp(spec.givenParseData, nil)

		if pd2.Source.pos != spec.expectedSrcPos {
			t.Errorf(
				"Expected source position %d, got %d.",
				spec.expectedSrcPos,
				pd2.Source.pos,
			)
		}
		if pd2.Result == nil {
			t.Fatalf("Expected a result.")
		}
		if pd2.Result.Pos != spec.expectedResult.Pos {
			t.Errorf(
				"Expected result position %d, got %d.",
				spec.expectedResult.Pos,
				pd2.Result.Pos,
			)
		}
		if pd2.Result.Text != spec.expectedResult.Text {
			t.Errorf(
				"Expected result text '%s', got '%s'.",
				spec.expectedResult.Text,
				pd2.Result.Text,
			)
		}
		if spec.expectedResult.Value == nil && pd2.Result.Value != nil {
			t.Errorf("Didn't expect a value but got '%#v'.", pd2.Result.Value)
		}
		if spec.expectedResult.Value != nil &&
			!reflect.DeepEqual(pd2.Result.Value, spec.expectedResult.Value) {

			t.Logf("The acutal value isn't equal to the expected one:")
			t.Errorf(
				"Expected value of type '%T', got '%T'.",
				spec.expectedResult.Value, pd2.Result.Value,
			)
			t.Errorf(
				"Expected value '%#v', got '%#v'.",
				spec.expectedResult.Value, pd2.Result.Value,
			)
		}

		if pd2.Result.ErrPos != spec.expectedResult.ErrPos {
			t.Errorf(
				"Expected result error position %d, got %d.",
				spec.expectedResult.ErrPos, pd2.Result.ErrPos,
			)
		}
		if pd2.Result.HasError() && spec.expectedErrCount <= 0 {
			t.Logf("Actual errors are: %s", printErrors(pd2.Result.Feedback))
			t.Fatalf("Expected no error but found at least one.")
		}
		if len(pd2.Result.Feedback) != spec.expectedErrCount {
			t.Logf("Actual errors are: %s", printErrors(pd2.Result.Feedback))
			t.Fatalf(
				"Expected %d errors, got %d.",
				spec.expectedErrCount, len(pd2.Result.Feedback),
			)
		}
		if spec.expectedErrCount > 0 &&
			pd2.Result.Feedback[spec.expectedErrCount-1].Msg.String() == "" {

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

func SemanticsTestOp(pd *ParseData, ctx interface{}) (*ParseData, interface{}) {
	pd.Result.Value = semanticTestValue
	return pd, nil
}
