package main

import (
	"bytes"
	"reflect"
	"testing"
)

func TestDiff(t *testing.T) {

	tests := []struct {
		diff string
		want Diff
	}{
		{
			diff: `diff --git main.go main.go
index e953e19..60856c9 100755
--- main.go
+++ main.go
@@ -1 +0,0 @@
-// hello
@@ -16,0 +16,1 @@ import (
+// Line Added
`,
			want: Diff{
				Added:   1,
				Removed: 1,
				Hunks: Hunks{
					hunkPair(1, 1, 0, 0, "@@ -1 +0,0 @@\n-// hello"),
					hunkPair(16, 0, 16, 1, "@@ -16,0 +16,1 @@ import (\n+// Line Added"),
				},
			},
		},
		{
			diff: `index e953e19..80cee70 100755
--- main.go
+++ main.go
@@ -1 +0,0 @@
-// hello
@@ -16,0 +16 @@ import (
+// Line added at middle of file
@@ -296,0 +298 @@ func bail(format string, args ...interface{}) {
+// Line added at end of file
`,
			want: Diff{
				Added:   2,
				Removed: 1,
				Hunks: Hunks{
					hunkPair(1, 1, 0, 0, "@@ -1 +0,0 @@\n-// hello"),
					hunkPair(16, 0, 16, 1, "@@ -16,0 +16 @@ import (\n+// Line added at middle of file"),
					hunkPair(296, 0, 298, 1, "@@ -296,0 +298 @@ func bail(format string, args ...interface{}) {\n+// Line added at end of file"),
				},
			},
		},
		{
			diff: `diff --git main.go main.go
index 4e3ecb5..b7d5fd7 100755
--- main.go
+++ main.go
@@ -1 +0,0 @@
-// hello
`,
			want: Diff{
				Added:   0,
				Removed: 1,
				Hunks: Hunks{
					hunkPair(1, 1, 0, 0, "@@ -1 +0,0 @@\n-// hello"),
				},
			},
		},
	}

	for i, tt := range tests {
		b := bytes.NewReader([]byte(tt.diff))
		got, err := NewDiff(b)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("tests[%d] failed\nwant: %s\n got: %s", i, &tt.want, &got)
		}
	}
}
