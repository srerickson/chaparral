package diff_test

import (
	"reflect"
	"testing"

	"github.com/srerickson/chaparral/internal/diff"
)

func TestDiff(t *testing.T) {
	tests := map[string]struct {
		a      map[string]string
		b      map[string]string
		result diff.Result
	}{
		"same_empty": {
			a:      map[string]string{},
			b:      map[string]string{},
			result: diff.Result{},
		},
		"same_simple": {
			a: map[string]string{
				"file":      "abc",
				"dir/file2": "abc",
			},
			b: map[string]string{
				"file":      "abc",
				"dir/file2": "abc",
			},
			result: diff.Result{},
		},
		"1_added": {
			a: map[string]string{
				"file": "abc",
			},
			b: map[string]string{
				"file":      "abc",
				"dir/file2": "abc",
			},
			result: diff.Result{
				Added: []string{"dir/file2"},
			},
		},
		"1_removed": {
			a: map[string]string{
				"file":      "abc",
				"dir/file2": "abc",
			},
			b: map[string]string{
				"file": "abc",
			},
			result: diff.Result{
				Removed: []string{"dir/file2"},
			},
		},
		"1_modified": {
			a: map[string]string{
				"file":      "abc",
				"dir/file2": "abc1",
			},
			b: map[string]string{
				"file":      "abc",
				"dir/file2": "abc2",
			},
			result: diff.Result{
				Modified: []string{"dir/file2"},
			},
		},
		"1_renamed": {
			a: map[string]string{
				"file":      "abc",
				"dir/file2": "abc2",
			},
			b: map[string]string{
				"file":      "abc",
				"dir/file3": "abc2",
			},
			result: diff.Result{
				Renamed: map[string]string{"dir/file2": "dir/file3"},
			},
		},
		"2_added_renamed": {
			a: map[string]string{
				"file":      "abc",
				"dir/file2": "abc2",
			},
			b: map[string]string{
				"file":      "abc",
				"dir/file3": "abc2",
				"dir/file4": "abc3",
			},
			result: diff.Result{
				Added:   []string{"dir/file4"},
				Renamed: map[string]string{"dir/file2": "dir/file3"},
			},
		},
		"2_removed_renamed": {
			a: map[string]string{
				"file1": "abc",
				"file2": "abc",
				"file3": "abc",
			},
			b: map[string]string{
				"file4": "abc",
			},
			result: diff.Result{
				Removed: []string{"file2", "file3"},
				Renamed: map[string]string{"file1": "file4"},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := diff.Diff(test.a, test.b)
			if err != nil {
				t.Fatal("diff.Diff() unexpected error", err)
			}
			if !reflect.DeepEqual(result, test.result) {
				t.Errorf("diff.Diff() unexpected result,got:\n%s\n\nexpect:\n%s\n", result, test.result)
			}
		})
	}
}
