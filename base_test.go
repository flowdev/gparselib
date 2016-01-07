package gparselib

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestWhere(t *testing.T) {
	src := &SourceData{name: "file1", content: "content\nline2\nline3\nand4\n",
		pos: 15, wherePrevNl: 13, whereLine: 3}

	Convey("Searching forward, ...", t, func() {
		Convey(`... should find a position in the same line.`, func() {
			where := where(src, 14)

			So(where, ShouldContainSubstring, "File '"+src.name+"'")
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

			So(where, ShouldContainSubstring, "File '"+src.name+"'")
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
		src := &SourceData{name: "empty", content: "",
			pos: 0, wherePrevNl: -1, whereLine: 1}

		Convey(`... should find start position.`, func() {
			where := where(src, 0)

			So(where, ShouldContainSubstring, "File '"+src.name+"'")
			So(where, ShouldContainSubstring, "line 1")
			So(where, ShouldContainSubstring, "column 1")
			So(where, ShouldEndWith, "\n")
		})
	})
}

func TestCreateUnmatchedResult(t *testing.T) {
	pd := NewParseData("file1", "content\nline2\nline3\nand4\n")
	pd.source.pos = 15
	pd.source.wherePrevNl = 13
	pd.source.whereLine = 3

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
			So(pd.Result.Feedback.Errors, ShouldNotBeNil)
			So(len(pd.Result.Feedback.Errors), ShouldEqual, 1)
			So(pd.Result.Feedback.Errors[0].Error(), ShouldEndWith, "\nBust.")
		})
	})
}

func TestCreateMatchedResult(t *testing.T) {
	pd := NewParseData("file1", "content\nline2\nline3\nand4\n")
	pd.source.pos = 15
	pd.source.wherePrevNl = 13
	pd.source.whereLine = 3

	createMatchedResult(pd, 3)

	Convey("Creating a matched result, ...", t, func() {
		Convey(`... should create result with text, no error position and no value.`, func() {
			So(pd.Result, ShouldNotBeNil)
			So(pd.Result.Pos, ShouldEqual, 15)
			So(pd.Result.ErrPos, ShouldEqual, -1)
			So(pd.Result.Text, ShouldEqual, "ine")
			So(pd.Result.Value, ShouldBeNil)
		})

		Convey(`... should give no error feedback.`, func() {
			So(pd.Result.Feedback.Errors, ShouldBeNil)
		})
	})
}

func TestMin(t *testing.T) {
	Convey("Testing min(a, b), ...", t, func() {
		So(min(1, 2), ShouldEqual, 1)
		So(min(2, 1), ShouldEqual, 1)
	})
}
func TestMax(t *testing.T) {
	Convey("Testing max(a, b), ...", t, func() {
		So(max(1, 2), ShouldEqual, 2)
		So(max(2, 1), ShouldEqual, 2)
	})
}
