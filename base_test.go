package gparselib

import (
	"strings"
	"testing"
)

func TestWhere(t *testing.T) {
	src := &SourceData{
		Name:        "file1",
		content:     "content\nline2\nline3\nand4\n",
		pos:         15,
		wherePrevNl: 13,
		whereLine:   3,
	}

	specs := []struct {
		givenSourceData *SourceData
		givenPosition   int
		expectedStrings []string
	}{
		{
			givenSourceData: src,
			givenPosition:   14,
			expectedStrings: []string{
				"File 'file1'",
				"line 3",
				"column 1",
				"\nline3\n",
			},
		}, {
			givenSourceData: src,
			givenPosition:   len(src.content) - 1,
			expectedStrings: []string{
				"File 'file1'",
				"line 4",
				"column 5",
				"\nand4\n",
			},
		}, {
			givenSourceData: src,
			givenPosition:   13,
			expectedStrings: []string{
				"File 'file1'",
				"line 2",
				"column 6",
				"\nline2\n",
			},
		}, {
			givenSourceData: src,
			givenPosition:   0,
			expectedStrings: []string{
				"File 'file1'",
				"line 1",
				"column 1",
				"\ncontent\n",
			},
		}, {
			givenSourceData: &SourceData{
				Name:        "empty",
				content:     "",
				pos:         0,
				wherePrevNl: -1,
				whereLine:   1,
			},
			givenPosition: 0,
			expectedStrings: []string{
				"File 'empty'",
				"line 1",
				"column 1",
				"\n\n",
			},
		},
	}

	for i, spec := range specs {
		t.Logf("Test run: %d", i)
		w := where(spec.givenSourceData, spec.givenPosition)
		for _, s := range spec.expectedStrings {
			if !strings.Contains(w, s) {
				t.Errorf(
					"Expected where string containing '%s', but got: '%s'",
					s,
					w,
				)
			}
		}
	}
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
