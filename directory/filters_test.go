package directory

import (
	"errors"
	"testing"

	"github.com/alvidir/filebrowser/file"
)

func TestAgregate(t *testing.T) {
	subject := NewDirectory(1)
	subject.files["a_file"], _ = file.NewFile("", "filename")
	subject.files["/another_file"], _ = file.NewFile("", "filename")
	subject.files["a_directory/a_file"], _ = file.NewFile("", "filename")
	subject.files["/a_directory/another_file"], _ = file.NewFile("", "filename")
	subject.files["/a_directory/another_dir/a_file"], _ = file.NewFile("", "filename")

	filterFn, err := newFilterByDirFn("/")
	if err != nil {
		t.Errorf("got error = %v, want = nil", err)
	}

	files, err := filterFiles(subject.Files(), []filterFileFn{filterFn})
	if err != nil {
		t.Errorf("got error = %v, want = nil", err)
	}

	agr, err := files.Agregate()
	if err != nil {
		t.Errorf("got error = %v, want = nil", err)
	}

	f, exists := agr["a_directory"]
	if !exists {
		t.Errorf("got exists = %v, want = %v", exists, true)
	}

	want := "3"
	if size, exists := f.Metadata()[file.MetadataSizeKey]; !exists || size != want {
		t.Errorf("got size = %v, want = %v", size, want)
	}
}

func TestFilterByDir(t *testing.T) {
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
			err: nil,
		},
		{
			name: "filter by another filename",
			path: "a_directory/a_file",
			want: []struct {
				path  string
				isDir bool
			}{},
			err: nil,
		},
		{
			name: "filter by non existing directory",
			path: "another_directory",
			want: []struct {
				path  string
				isDir bool
			}{},
			err: nil,
		},
	}

	for _, test := range tests {
		t := t
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			subject := NewDirectory(1)
			subject.files["a_file"], _ = file.NewFile("", "filename")
			subject.files["/another_file"], _ = file.NewFile("", "filename")
			subject.files["a_directory/a_file"], _ = file.NewFile("", "filename")
			subject.files["/a_directory/another_file"], _ = file.NewFile("", "filename")

			filterFn, err := newFilterByDirFn(test.path)
			if err != nil {
				t.Errorf("got error = %v, want = nil", err)
			}

			files, err := filterFiles(subject.Files(), []filterFileFn{filterFn})
			if test.err == nil && err != nil {
				t.Errorf("got error = %v, want = nil", err)
			} else if test.err != nil && !errors.Is(test.err, err) {
				t.Errorf("got error = %v, want = %v", err, test.err)
			}

			if len(test.want) != len(files) {
				t.Errorf("got files length = %v, want = %v", len(files), len(test.want))
			}

			for _, fwant := range test.want {
				var exists bool
				files.Range(func(p string, f *file.File) bool {
					if exists = p == fwant.path; !exists {
						return true
					}

					if isDir := f.Flags()&file.Directory != 0; isDir != fwant.isDir {
						t.Errorf("got file path = %v as directory = %v, want = %v", p, isDir, fwant.isDir)
					}

					return false

				})

				if !exists {
					t.Errorf("path = %v not found", fwant.path)
				}
			}
		})
	}
}
