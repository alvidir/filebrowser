package file

import (
	"errors"
	"testing"

	fb "github.com/alvidir/filebrowser"
)

func TestFileModel_NewFile(t *testing.T) {
	id := "id"
	filename := "filename"

	if file, err := NewFile(id, filename); err != nil {
		t.Errorf("got error = %v, want = %v", err, nil)
	} else if file.id != id {
		t.Errorf("got file.id = %v, want = %v", file.id, id)
	} else if file.name != filename {
		t.Errorf("got file.name = %v, want = %v", file.name, filename)
	} else if file.flags != 0 {
		t.Errorf("got file.flags = %v, want = %v", file.flags, 0)
	} else if file.metadata == nil {
		t.Errorf("got file.metadata = %v, want = %v", file.metadata, Metadata{})
	} else if file.permissions == nil {
		t.Errorf("got file.permissions = %v, want = %v", file.permissions, Permissions{})
	} else if file.data == nil || len(file.data) > 0 {
		t.Errorf("got file.data = %v, want = %v", file.data, []byte{})
	}

	filename = "/filename"
	if _, err := NewFile(id, filename); !errors.Is(err, fb.ErrInvalidFormat) {
		t.Errorf("got error = %v, want = %v", err, fb.ErrInvalidFormat)
	}

	filename = "file/name"
	if _, err := NewFile(id, filename); !errors.Is(err, fb.ErrInvalidFormat) {
		t.Errorf("got error = %v, want = %v", err, fb.ErrInvalidFormat)
	}

	filename = "filename/"
	if _, err := NewFile(id, filename); !errors.Is(err, fb.ErrInvalidFormat) {
		t.Errorf("got error = %v, want = %v", err, fb.ErrInvalidFormat)
	}
}
