package directory

import (
	"fmt"
	"testing"

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

	path := "/path/to/filename"
	if got := dir.AddFile(f, path); got != path {
		t.Errorf("got final path = %s, want = %v", got, path)
	}

	if got := dir.files[path]; got.Id() != f.Id() {
		t.Errorf("got file id = %v, want = %v", got, f.Id())
	}

	want1 := fmt.Sprintf("%s_1", path)
	if got := dir.AddFile(f, path); got != want1 {
		t.Errorf("got final path = %s, want = %v", got, want1)
	}

	want2 := fmt.Sprintf("%s_2", path)
	if got := dir.AddFile(f, path); got != want2 {
		t.Errorf("got final path = %s, want = %v", got, want2)
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
		"/path/to/filename",
		"/path/to/filename",
		"/antoher/file",
	}

	for _, path := range want {
		dir.AddFile(f, path)
	}

	want[1] = "/path/to/filename_1"

	got := dir.files
	if len(got) != len(want) {
		t.Errorf("got len = %v, want = %v", len(got), len(want))
	}

	for _, path := range want {
		if got, exists := got[path]; !exists || got.Id() != f.Id() {
			t.Errorf("got file id = %v, want = %v", got, f.Id())
		}
	}
}

func TestAgregateFiles(t *testing.T) {
	tests := []struct {
		name  string
		path  string
		files []string
		want  []string
	}{
		{
			name: "get root files",
			path: "/",
			files: []string{
				"/a_file",
				"/another_file",
				"/a_directory/a_file",
				"/a_directory/another_file",
				"/another_directory/a_file",
			},
			want: []string{
				"/a_file",
				"/another_file",
				"/a_directory",
				"/another_directory",
			},
		},
		{
			name: "get deeper directory files",
			path: "/a_directory",
			files: []string{
				"/a_file",
				"/another_file",
				"/a_directory/a_file",
				"/a_directory/another_file",
				"/a_directory_here/a_file",
			},
			want: []string{
				"/a_directory/a_file",
				"/a_directory/another_file",
			},
		},
	}

	for _, test := range tests {
		t := t
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			dir := &Directory{
				id:     "test",
				userId: 999,
				files:  make(map[string]*file.File),
			}

			for _, fileName := range test.files {
				dir.files[fileName], _ = file.NewFile("test", "filename")
			}

			got := dir.AggregateFiles(test.path)
			if len(test.want) != len(got) {
				t.Fatalf("got files = %v, want = %v", got, test.want)
			}

			for _, expectedPath := range test.want {
				if _, exists := got[expectedPath]; !exists {
					t.Errorf("got files = %v, want = %v", got, expectedPath)
				}
			}
		})
	}
}
