package gparselib

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestParseLiteral(t *testing.T) {
	p := NewParseLiteral(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd }, "func")

	runTest(t, p, newData("no match", 0, " func\n"), newResult(0, "", nil, 0), 0, 1)
	runTest(t, p, newData("empty", 0, ""), newResult(0, "", nil, 0), 0, 1)
	runTest(t, p, newData("simple", 0, "func"), newResult(0, "func", nil, -1), 4, 0)
	runTest(t, p, newData("simple 2", 0, "func 123"), newResult(0, "func", nil, -1), 4, 0)
	runTest(t, p, newData("simple 3", 2, "12func345"), newResult(2, "func", nil, -1), 6, 0)
}

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

func TestParseBlockComment(t *testing.T) {
	p := NewParseBlockComment(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd }, "/*", "*/")

	runTest(t, p, newData("no match", 0, " 123 "), newResult(0, "", nil, 0), 0, 1)
	runTest(t, p, newData("empty", 0, ""), newResult(0, "", nil, 0), 0, 1)
	runTest(t, p, newData("simple", 0, "/* 123 */"), newResult(0, "/* 123 */", "", -1), 9, 0)
	runTest(t, p, newData("simple 2", 4, "abcd/* 123 */"), newResult(4, "/* 123 */", "", -1), 13, 0)
	runTest(t, p, newData("simple 3", 2, "ab/* 123 */cdefg"), newResult(2, "/* 123 */", "", -1), 11, 0)
	runTest(t, p, newData("nested block comments aren't supported!!!", 2, "ab/* 1 /* 2 */ 3 */cdefg"),
		newResult(2, "/* 1 /* 2 */", "", -1), 14, 0)
	runTest(t, p, newData("comment not closed", 4, "abcd/* 123 "), newResult(4, "", nil, 6), 6, 1)
	runTest(t, p, newData("comment in single qoute string", 0, `/* 1'2\'*/'3 */`),
		newResult(0, `/* 1'2\'*/'3 */`, "", -1), 15, 0)
	runTest(t, p, newData("comment in double qoute string", 0, `/* 1"2\"*/"3 */`),
		newResult(0, `/* 1"2\"*/"3 */`, "", -1), 15, 0)
	runTest(t, p, newData("comment in backqoute string", 0, "/* 1`2*/\\`*/`3 */"),
		newResult(0, "/* 1`2*/\\`*/", "", -1), 12, 0)
}

const semanticTestValue = "Semantic test!!!"

func TestParseSemantics(t *testing.T) {
	p := NewParseLiteral(func(data interface{}) *ParseData { return data.(*ParseData) },
		func(data interface{}, pd *ParseData) interface{} { return pd }, "func")
	s := &SemanticsTestOp{}
	p.SetSemOutPort(s.InPort)
	s.SetOutPort(p.SemInPort)

	runTest(t, p, newData("no match", 0, " func\n"), newResult(0, "", nil, 0), 0, 1)
	runTest(t, p, newData("simple", 0, "func"), newResult(0, "func", semanticTestValue, -1), 4, 0)
}

func newResult(pos int, text string, value interface{}, errPos int) *ParseResult {
	return &ParseResult{Pos: pos, Text: text, Value: value, ErrPos: errPos}
}
func newData(srcName string, srcPos int, srcContent string) *ParseData {
	pd := NewParseData(srcName, srcContent)
	pd.source.pos = srcPos
	return pd
}

func runTest(t *testing.T, sp SimpleParseOp, pd *ParseData, er *ParseResult, newSrcPos int, errCount int) {
	var pd2 *ParseData
	sp.SetOutPort(func(data interface{}) { pd2 = data.(*ParseData) })
	sp.InPort(pd)

	Convey("Simple parsing '"+pd.source.name+"', ...", t, func() {
		Convey(`... should give the right source position.`, func() {
			So(pd2.source.pos, ShouldEqual, newSrcPos)
		})
		Convey(`... should create a result.`, func() {
			So(pd2.Result, ShouldNotBeNil)
		})
		valueTest(pd2.Result.Value, er.Value)
		Convey(`... should give the right error position.`, func() {
			So(pd2.Result.ErrPos, ShouldEqual, er.ErrPos)
		})
		Convey(`... should give the right result position.`, func() {
			So(pd2.Result.Pos, ShouldEqual, er.Pos)
		})
		Convey(`... should give the right result text.`, func() {
			So(pd2.Result.Text, ShouldEqual, er.Text)
		})
		Convey(`... should give the right errors.`, func() {
			if errCount <= 0 {
				So(pd2.Result.Feedback.Errors, ShouldBeNil)
			} else {
				So(len(pd2.Result.Feedback.Errors), ShouldEqual, errCount)
				So(pd2.Result.Feedback.Errors[errCount-1].Error(), ShouldNotBeNil)
			}
		})
	})
}
func valueTest(actual, expected interface{}) {
	if expected == nil {
		Convey(`... should create no value.`, func() {
			So(actual, ShouldBeNil)
		})
	} else {
		Convey(`... should create the right value.`, func() {
			So(fmt.Sprintf("%T", actual), ShouldEqual, fmt.Sprintf("%T", expected))
			So(fmt.Sprintf("%#v", actual), ShouldEqual, fmt.Sprintf("%#v", expected))
		})
	}
}

type SemanticsTestOp struct {
	outPort func(interface{})
}

func (s *SemanticsTestOp) InPort(data interface{}) {
	p := data.(*ParseData)
	p.Result.Value = semanticTestValue
	s.outPort(p)
}
func (p *SemanticsTestOp) SetOutPort(outPort func(interface{})) {
	p.outPort = outPort
}
