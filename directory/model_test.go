package directory

import (
	"fmt"
	"testing"

	"github.com/alvidir/filebrowser/file"
)

func TestDirectoryModel_AddFile(t *testing.T) {
	subject := NewDirectory(0)
	dest := "path/to/filename"

	f := file.NewFile("111", "test", nil)

	want := dest
	if got := subject.AddFile(f, dest); got != want {
		t.Errorf("got available destination = %s, want = %v", got, want)
	}

	if got := subject.files[want]; got.Id() != f.Id() {
		t.Errorf("got file id = %s, want = %v", got.Id(), f.Id())
	} else if got := got.Name(); got != f.Name() {
		t.Errorf("got file name = %s, want = %v", got, f.Name())
	}

	want = fmt.Sprintf("%s (1)", dest)
	if got := subject.AddFile(f, dest); got != want {
		t.Errorf("got available destination = %s, want = %v", got, want)
	}

	want = fmt.Sprintf("%s (2)", dest)
	if got := subject.AddFile(f, dest); got != want {
		t.Errorf("got available destination = %s, want = %v", got, want)
	}
}
