package gparselib

import (
	"errors"
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

func TestGetFeedback(t *testing.T) {
	specs := []struct {
		name           string
		givenParseData *ParseData
		expectedInfo   string
		expectedError  error
	}{
		{
			name: "no feedback",
			givenParseData: &ParseData{
				Result: &ParseResult{
					Feedback: nil,
				},
			},
			expectedInfo:  "",
			expectedError: nil,
		}, {
			name: "empty feedback",
			givenParseData: &ParseData{
				Result: &ParseResult{
					Feedback: []*FeedbackItem{},
				},
			},
			expectedInfo:  "",
			expectedError: nil,
		}, {
			name: "info feedback",
			givenParseData: &ParseData{
				Result: &ParseResult{
					Feedback: []*FeedbackItem{
						{
							Kind: FeedbackInfo,
							Msg:  &parseMessage{msg: "msg 1"},
						},
					},
				},
			},
			expectedInfo:  "INFO: msg 1.",
			expectedError: nil,
		}, {
			name: "warning feedback",
			givenParseData: &ParseData{
				Result: &ParseResult{
					Feedback: []*FeedbackItem{
						{
							Kind: FeedbackWarning,
							Msg:  &parseMessage{msg: "msg 1"},
						},
					},
				},
			},
			expectedInfo:  "WARNING: msg 1.",
			expectedError: nil,
		}, {
			name: "error feedback",
			givenParseData: &ParseData{
				Result: &ParseResult{
					Feedback: []*FeedbackItem{
						{
							Kind: FeedbackError,
							Msg:  &parseError{myErr: "err 1"},
						},
					},
				},
			},
			expectedInfo:  "",
			expectedError: &parseError{myErr: "ERROR: err 1"},
		}, {
			name: "all feedback",
			givenParseData: &ParseData{
				Result: &ParseResult{
					Feedback: []*FeedbackItem{
						{
							Kind: FeedbackError,
							Msg:  &parseError{myErr: "err 2"},
						}, {
							Kind: FeedbackWarning,
							Msg:  &parseMessage{msg: "warn 2"},
						}, {
							Kind: FeedbackInfo,
							Msg:  &parseMessage{msg: "info 2"},
						},
					},
				},
			},
			expectedInfo:  "WARNING: warn 2.\nINFO: info 2.",
			expectedError: &parseError{myErr: "ERROR: err 2"},
		}, {
			name: "multi error feedback",
			givenParseData: &ParseData{
				Result: &ParseResult{
					Feedback: []*FeedbackItem{
						{
							Kind: FeedbackError,
							Msg:  &parseError{myErr: "err 3"},
						}, {
							Kind: FeedbackError,
							Msg:  &parseError{myErr: "err 4"},
						},
					},
				},
			},
			expectedInfo:  "",
			expectedError: errors.New("ERROR: err 3.\nERROR: err 4."),
		},
	}

	for _, spec := range specs {
		t.Run(spec.name, func(tt *testing.T) {
			actualInfo, actualError := spec.givenParseData.GetFeedback()

			if actualInfo != spec.expectedInfo {
				t.Errorf("expected feedback msg %q, got: %q", spec.expectedInfo, actualInfo)
			}
			if (actualError != nil && spec.expectedError == nil) ||
				(actualError == nil && spec.expectedError != nil) ||
				(actualError != nil && spec.expectedError != nil && actualError.Error() != spec.expectedError.Error()) {
				t.Errorf(`expected feedback error "%v", got: "%v"`, spec.expectedError, actualError)
			}
		})
	}
}
