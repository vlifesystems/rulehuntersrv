package html

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestGenReportFilename(t *testing.T) {
	wwwDir := "/var/wwww"
	cases := []struct {
		stamp        time.Time
		title        string
		wantFilename string
	}{
		{time.Date(2009, time.November, 10, 22, 19, 18, 17, time.UTC),
			"This could be very interesting",
			filepath.Join(wwwDir, "reports", "2009", "11", "10",
				fmt.Sprintf("%s_this-could-be-very-interesting",
					genStampMagicString(
						time.Date(2009, time.November, 10, 22, 19, 18, 17, time.UTC),
					),
				),
				"index.html")},
	}
	for _, c := range cases {
		got := genReportFilename(wwwDir, c.stamp, c.title)
		if got != c.wantFilename {
			t.Errorf("genReportFilename(%s, %s) got: %s, want: %s",
				c.stamp, c.title, got, c.wantFilename)
		}
	}
}

func TestGenReportURLDir(t *testing.T) {
	cases := []struct {
		stamp   time.Time
		title   string
		wantDir string
	}{
		{time.Date(2009, time.November, 10, 22, 19, 18, 17, time.UTC),
			"This could be very interesting",
			fmt.Sprintf("/reports/2009/11/10/%s_this-could-be-very-interesting",
				genStampMagicString(
					time.Date(2009, time.November, 10, 22, 19, 18, 17, time.UTC),
				),
			),
		},
	}
	for _, c := range cases {
		got := genReportURLDir(c.stamp, c.title)
		if got != c.wantDir {
			t.Errorf("genReportFilename(%s, %s) got: %s, want: %s",
				c.stamp, c.title, got, c.wantDir)
		}
	}
}

func TestMakeReportURLDir(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "html_test")
	if err != nil {
		t.Errorf("TempDir() err: %s", err)
		return
	}
	defer os.RemoveAll(tempDir)
	cases := []struct {
		stamp         time.Time
		title         string
		wantDirExists string
		wantURLDir    string
	}{
		{time.Date(2009, time.November, 10, 22, 19, 18, 17, time.UTC),
			"This could be very interesting",
			filepath.Join(tempDir, "reports", "2009", "11", "10",
				fmt.Sprintf("%s_this-could-be-very-interesting",
					genStampMagicString(
						time.Date(2009, time.November, 10, 22, 19, 18, 17, time.UTC),
					),
				),
			),
			fmt.Sprintf("/reports/2009/11/10/%s_this-could-be-very-interesting",
				genStampMagicString(
					time.Date(2009, time.November, 10, 22, 19, 18, 17, time.UTC),
				),
			),
		},
	}
	for _, c := range cases {
		got, err := makeReportURLDir(tempDir, c.stamp, c.title)
		if err != nil {
			t.Errorf("makeReportURLDir(%s, %s, %s) err: %s",
				tempDir, c.stamp, c.title, err)
		}
		if got != c.wantURLDir {
			t.Errorf("makeReportURLDir(%s, %s, %s) got: %s, want: %s",
				tempDir, c.stamp, c.title, got, c.wantURLDir)
		}
		if !dirExists(c.wantDirExists) {
			t.Errorf("makeReportURLDir(%s, %s, %s)  - directory doesn't exist: %s",
				tempDir, c.stamp, c.title, c.wantDirExists)
		}
	}
}

func TestEscapeString(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"This is a TITLE",
			"this-is-a-title"},
		{"  hello how are % you423 33  today __ --",
			"hello-how-are-you423-33-today"},
		{"--  hello how are %^& you423 33  today __ --",
			"hello-how-are-you423-33-today"},
		{"hello((_ how are % you423 33  today",
			"hello-how-are-you423-33-today"},
		{"", ""},
	}
	for _, c := range cases {
		got := escapeString(c.in)
		if got != c.want {
			t.Errorf("escapeString(%s) got: %s, want: %s", c.in, got, c.want)
		}
	}
}

func TestGenStampMagicString(t *testing.T) {
	cases := []struct {
		in       time.Time
		wantDiff uint64
	}{
		{time.Date(2009, time.November, 10, 22, 19, 18, 200, time.UTC), 0},
		{time.Date(2009, time.November, 11, 22, 19, 18, 200, time.UTC), 0},
		{time.Date(2009, time.December, 11, 22, 19, 18, 200, time.UTC), 0},
		{time.Date(2010, time.December, 11, 22, 19, 18, 200, time.UTC), 0},
		{time.Date(2009, time.November, 10, 22, 19, 19, 17, time.UTC), 1},
		{time.Date(2009, time.November, 10, 22, 19, 29, 17, time.UTC), 11},
		{time.Date(2009, time.November, 10, 22, 20, 18, 17, time.UTC), 60},
		{time.Date(2009, time.November, 10, 23, 19, 18, 17, time.UTC), 3600},
	}

	initStamp := time.Date(2009, time.November, 10, 22, 19, 18, 17, time.UTC)
	initMagicStr := genStampMagicString(initStamp)

	initMagicNum, err := strconv.ParseUint(initMagicStr, 36, 64)
	if err != nil {
		t.Errorf("ParseUint(%s, 36, 64) err: %s", initMagicStr, err)
		return
	}

	for _, c := range cases {
		magicStr := genStampMagicString(c.in)
		magicNum, err := strconv.ParseUint(magicStr, 36, 64)
		if err != nil {
			t.Errorf("ParseUint(%s, 36, 64) err: %s", magicStr, err)
			return
		}
		diff := magicNum - initMagicNum
		if diff != c.wantDiff {
			t.Errorf("diff != wantDiff for stamp: %s got: %d, want: %d",
				c.in, diff, c.wantDiff)
		}
	}
}

func dirExists(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fi.IsDir()
}
