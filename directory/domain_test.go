package directory

import (
	"errors"
	"fmt"
	"testing"

	fb "github.com/alvidir/filebrowser"
	"github.com/alvidir/filebrowser/file"
)

func TestNewDirectory(t *testing.T) {
	var want int32 = 999
	dir := NewDirectory(want)

	if got := dir.id; got != "" {
		t.Errorf("got id = %v, want = %v", got, "''")
	}

	if got := dir.userId; got != want {
		t.Errorf("got user_id = %v, want = %v", got, want)
	}
}

func TestAddFile(t *testing.T) {
	dir := NewDirectory(999)

	f, err := file.NewFile("111", "test")
	if err != nil {
		t.Errorf("got error = %v, want = %v", err, nil)
	}

	want := "path/to/filename"
	if got := dir.AddFile(f, want); got != want {
		t.Errorf("got final path = %s, want = %v", got, want)
	}

	if got := dir.files[want]; got.Id() != f.Id() {
		t.Errorf("got file id = %v, want = %v", got, f.Id())
	}

	want = fmt.Sprintf("%s (1)", want)
	if got := dir.AddFile(f, want); got != want {
		t.Errorf("got final path = %s, want = %v", got, want)
	}

	want = fmt.Sprintf("%s (2)", want)
	if got := dir.AddFile(f, want); got != want {
		t.Errorf("got final path = %s, want = %v", got, want)
	}
}

func TestRemoveFile(t *testing.T) {
	dir := NewDirectory(999)

	f, err := file.NewFile("111", "test")
	if err != nil {
		t.Errorf("got error = %v, want = %v", err, nil)
	}

	path := "path/to/filename"
	dir.AddFile(f, path)

	dir.RemoveFile(f)
	if got := len(dir.files); got != 0 {
		t.Errorf("got files len = %v, want = %v", got, 0)
	}
}

func TestFiles(t *testing.T) {
	dir := NewDirectory(999)

	f, err := file.NewFile("111", "test")
	if err != nil {
		t.Errorf("got error = %v, want = %v", err, nil)
	}

	want := []string{
		"path/to/filename",
		"path/to/filename",
		"antoher/file",
	}

	for _, path := range want {
		dir.AddFile(f, path)
	}

	want[1] = "path/to/filename (1)"

	got := dir.Files()
	if len(got) != len(want) {
		t.Errorf("got len = %v, want = %v", len(got), len(want))
	}

	for _, path := range want {
		if got, exists := got[path]; !exists || got.Id() != f.Id() {
			t.Errorf("got file id = %v, want = %v", got, f.Id())
		}
	}
}

func TestFilterByPath(t *testing.T) {
	subject := NewDirectory(1)
	subject.files["a_file"], _ = file.NewFile("", "filename")
	subject.files["/another_file"], _ = file.NewFile("", "filename")
	subject.files["a_directory/a_file"], _ = file.NewFile("", "filename")
	subject.files["/a_directory/another_file"], _ = file.NewFile("", "filename")

	tests := []struct {
		name string
		path string
		want []struct {
			path  string
			isDir bool
		}
		err error
	}{
		{
			name: "filter by root",
			path: "/",
			want: []struct {
				path  string
				isDir bool
			}{
				{path: "a_file", isDir: false},
				{path: "another_file", isDir: false},
				{path: "a_directory", isDir: true},
			},
			err: nil,
		},
		{
			name: "filter by directory",
			path: "a_directory",
			want: []struct {
				path  string
				isDir bool
			}{
				{path: "a_file", isDir: false},
				{path: "another_file", isDir: false},
			},
			err: nil,
		},
		{
			name: "filter by filename",
			path: "a_file",
			want: []struct {
				path  string
				isDir bool
			}{},
			err: fb.ErrNotFound,
		},
		{
			name: "filter by another filename",
			path: "a_directory/a_file",
			want: []struct {
				path  string
				isDir bool
			}{},
			err: fb.ErrNotFound,
		},
		{
			name: "filter by non existing name",
			path: "another_directory",
			want: []struct {
				path  string
				isDir bool
			}{},
			err: fb.ErrNotFound,
		},
	}

	for _, test := range tests {
		t := t
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			files, err := subject.FilesByPath(test.path)
			if test.err == nil && err != nil {
				t.Errorf("got error = %v, want = nil", err)
			} else if test.err != nil && !errors.Is(test.err, err) {
				t.Errorf("got error = %v, want = %v", err, test.err)
			}

			if len(test.want) != len(files) {
				t.Errorf("got files length = %v, want = %v", len(test.want), len(files))
			}

			for _, fwant := range test.want {
				var exists bool
				for p, f := range files {
					if exists = p == fwant.path; !exists {
						continue
					}

					if isDir := f.Flags()&file.Directory != 0; isDir != fwant.isDir {
						t.Errorf("got file path = %v as directory = %v, want = %v", p, isDir, fwant.isDir)
					}

					break
				}

				if !exists {
					t.Errorf("path = %v not found", fwant.path)
				}
			}
		})
	}
}
