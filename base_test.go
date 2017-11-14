package gparselib

import (
	"strings"
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
	specs := []struct {
		givenParseData       *ParseData
		givenErrOffset       int
		givenErrMsg          string
		expectedResultPos    int
		expectedResultErrPos int
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
			givenErrOffset:       0,
			givenErrMsg:          "Bust1",
			expectedResultPos:    15,
			expectedResultErrPos: 15,
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
			givenErrOffset:       3,
			givenErrMsg:          "Bust2",
			expectedResultPos:    15,
			expectedResultErrPos: 18,
		},
	}

	for i, spec := range specs {
		t.Logf("Test run: %d", i)
		pd := spec.givenParseData
		createUnmatchedResult(pd, spec.givenErrOffset, spec.givenErrMsg, nil)
		if pd.Result == nil {
			t.Errorf("The result shouldn't be nil but it is!")
			break
		}
		if pd.Result.Pos != spec.expectedResultPos {
			t.Errorf(
				"The result position should be %d, but is %d.",
				spec.expectedResultPos,
				spec.givenParseData.Result.Pos,
			)
		}
		if pd.Result.ErrPos != spec.expectedResultErrPos {
			t.Errorf(
				"The result error position should be %d, but is %d.",
				spec.expectedResultErrPos,
				pd.Result.ErrPos,
			)
		}
		if pd.Result.Text != "" {
			t.Errorf(
				"The result text should be empty, but is '%s'.",
				pd.Result.Text,
			)
		}
		if pd.Result.Value != nil {
			t.Errorf(
				"The result value should be nil but it is: %#v",
				pd.Result.Value,
			)
		}
		if !pd.Result.HasError() {
			t.Errorf("The result should contain an error, but it doesn't!")
		}
		if len(pd.Result.Feedback) != 1 {
			t.Errorf(
				"Expected length of feedback to be 1, but it is: %d",
				len(pd.Result.Feedback),
			)
		}
		if !strings.Contains(pd.Result.Feedback[0].String(), spec.givenErrMsg) {
			t.Errorf(
				"Expected error message containing '%s', but got: '%s'",
				spec.givenErrMsg,
				pd.Result.Feedback[0].String(),
			)
		}
	}
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

	for i, spec := range specs {
		t.Logf("Test run: %d", i)
		pd := spec.givenParseData
		createMatchedResult(pd, spec.givenN)
		if pd.Result == nil {
			t.Errorf("The result shouldn't be nil but it is!")
			break
		}
		if pd.Result.Pos != spec.expectedResultPos {
			t.Errorf(
				"The result position should be %d, but is %d.",
				spec.expectedResultPos,
				spec.givenParseData.Result.Pos,
			)
		}
		if pd.Result.ErrPos != spec.expectedResultErrPos {
			t.Errorf(
				"The result error position should be %d, but is %d.",
				spec.expectedResultErrPos,
				pd.Result.ErrPos,
			)
		}
		if pd.Result.Text != spec.expectedResultText {
			t.Errorf(
				"The result text should be %s, but is %s.",
				spec.expectedResultText,
				pd.Result.Text,
			)
		}
		if pd.Result.Value != nil {
			t.Errorf(
				"The result value should be nil but it is: %#v",
				pd.Result.Value,
			)
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
