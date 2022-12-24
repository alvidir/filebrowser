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

	want1 := fmt.Sprintf("%s (1)", path)
	if got := dir.AddFile(f, path); got != want1 {
		t.Errorf("got final path = %s, want = %v", got, want1)
	}

	want2 := fmt.Sprintf("%s (2)", path)
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

	want[1] = "/path/to/filename (1)"

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
