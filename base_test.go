package gparselib

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestWhere(t *testing.T) {
	src := &SourceData{Name: "file1", content: "content\nline2\nline3\nand4\n",
		pos: 15, wherePrevNl: 13, whereLine: 3}

	Convey("Searching forward, ...", t, func() {
		Convey(`... should find a position in the same line.`, func() {
			where := where(src, 14)

			So(where, ShouldContainSubstring, "File '"+src.Name+"'")
			So(where, ShouldContainSubstring, "line 3")
			So(where, ShouldContainSubstring, "column 1")
			So(where, ShouldEndWith, "\nline3\n")
		})

		Convey(`... should find position at the end.`, func() {
			where := where(src, len(src.content)-1)

			So(where, ShouldContainSubstring, "line 4")
			So(where, ShouldContainSubstring, "column 5")
			So(where, ShouldEndWith, "\nand4\n")
		})
	})

	Convey("Searching backward, ...", t, func() {
		Convey(`... should find a position in the previous line.`, func() {
			where := where(src, 13)

			So(where, ShouldContainSubstring, "File '"+src.Name+"'")
			So(where, ShouldContainSubstring, "line 2")
			So(where, ShouldContainSubstring, "column 6")
			So(where, ShouldEndWith, "\nline2\n")
		})

		Convey(`... should find start position.`, func() {
			where := where(src, 0)

			So(where, ShouldContainSubstring, "line 1")
			So(where, ShouldContainSubstring, "column 1")
			So(where, ShouldEndWith, "\ncontent\n")
		})
	})

	Convey("Searching in empty content, ...", t, func() {
		src := &SourceData{Name: "empty", content: "",
			pos: 0, wherePrevNl: -1, whereLine: 1}

		Convey(`... should find start position.`, func() {
			where := where(src, 0)

			So(where, ShouldContainSubstring, "File '"+src.Name+"'")
			So(where, ShouldContainSubstring, "line 1")
			So(where, ShouldContainSubstring, "column 1")
			So(where, ShouldEndWith, "\n")
		})
	})
}

func TestCreateUnmatchedResult(t *testing.T) {
	pd := NewParseData("file1", "content\nline2\nline3\nand4\n")
	pd.Source.pos = 15
	pd.Source.wherePrevNl = 13
	pd.Source.whereLine = 3

	createUnmatchedResult(pd, 0, "Bust", nil)

	Convey("Creating an unmatched result, ...", t, func() {
		Convey(`... should create result with error position, empty text and no value.`, func() {
			So(pd.Result, ShouldNotBeNil)
			So(pd.Result.Pos, ShouldEqual, 15)
			So(pd.Result.ErrPos, ShouldEqual, 15)
			So(pd.Result.Text, ShouldBeEmpty)
			So(pd.Result.Value, ShouldBeNil)
		})

		Convey(`... should give error feedback.`, func() {
			So(pd.Result.HasError(), ShouldBeTrue)
			So(len(pd.Result.Feedback), ShouldEqual, 1)
			So(pd.Result.Feedback[0].String(), ShouldEndWith, "\nBust.")
		})
	})
}

func TestCreateMatchedResult(t *testing.T) {
	specs := []struct {
		givenParseData       *ParseData
		givenN               int
		expectedResultPos    int
		expectedResultErrPos int
		expectedResultText   string
	}{
		{
			givenParseData: &ParseData{
				Source: SourceData{
					Name:        "file1",
					content:     "content\nline2\nline3\nand4\n",
					pos:         15,
					wherePrevNl: 13,
					whereLine:   3,
				},
			},
			givenN:               0,
			expectedResultPos:    15,
			expectedResultErrPos: -1,
			expectedResultText:   "",
		}, {
			givenParseData: &ParseData{
				Source: SourceData{
					Name:        "file1",
					content:     "content\nline2\nline3\nand4\n",
					pos:         15,
					wherePrevNl: 13,
					whereLine:   3,
				},
			},
			givenN:               4,
			expectedResultPos:    15,
			expectedResultErrPos: -1,
			expectedResultText:   "ine3",
		},
	}

	for _, spec := range specs {
		createMatchedResult(spec.givenParseData, spec.givenN)
		if spec.givenParseData.Result == nil {
			t.Errorf("The result for result size %d shouldn't be nil but it is!", spec.givenN)
			break
		}
		if spec.givenParseData.Result.Pos != spec.expectedResultPos {
			t.Errorf("The result position should be %d, but is %d.", spec.expectedResultPos, spec.givenParseData.Result.Pos)
		}
		if spec.givenParseData.Result.ErrPos != spec.expectedResultErrPos {
			t.Errorf("The result error position should be %d, but is %d.", spec.expectedResultErrPos, spec.givenParseData.Result.ErrPos)
		}
		if spec.givenParseData.Result.Text != spec.expectedResultText {
			t.Errorf("The result text should be %s, but is %s.", spec.expectedResultText, spec.givenParseData.Result.Text)
		}
		if spec.givenParseData.Result.Value != nil {
			t.Errorf("The result value for result size %d should be nil but it is: %#v", spec.givenN, spec.givenParseData.Result.Value)
		}
	}
}

func TestMinMax(t *testing.T) {
	specs := []struct {
		givenA         int
		givenB         int
		expectedResult int
		actualResult   int
		name           string
	}{
		{givenA: 1, givenB: 2, expectedResult: 2, actualResult: max(1, 2), name: "max"},
		{givenA: 2, givenB: 1, expectedResult: 2, actualResult: max(2, 1), name: "max"},
		{givenA: 1, givenB: 2, expectedResult: 1, actualResult: min(1, 2), name: "min"},
		{givenA: 2, givenB: 1, expectedResult: 1, actualResult: min(2, 1), name: "min"},
	}
	for _, spec := range specs {
		if spec.actualResult != spec.expectedResult {
			t.Errorf("%s(%d, %d) != %d", spec.name, spec.givenA, spec.givenB, spec.expectedResult)
		}
	}
}
