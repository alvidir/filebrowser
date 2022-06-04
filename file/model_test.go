package file

import (
	"errors"
	"testing"

	fb "github.com/alvidir/filebrowser"
)

func TestNewFile(t *testing.T) {
	id := "id"
	filename := "filename"

	if file, err := NewFile(id, filename); err != nil {
		t.Errorf("got error = %v, want = %v", err, nil)
	} else if file.id != id {
		t.Errorf("got id = %v, want = %v", file.id, id)
	} else if file.name != filename {
		t.Errorf("got name = %v, want = %v", file.name, filename)
	} else if file.flags != 0 {
		t.Errorf("got flags = %v, want = %v", file.flags, 0)
	} else if file.metadata == nil {
		t.Errorf("got metadata = %v, want = %v", file.metadata, Metadata{})
	} else if file.permissions == nil {
		t.Errorf("got permissions = %v, want = %v", file.permissions, Permissions{})
	} else if file.data == nil || len(file.data) > 0 {
		t.Errorf("got data = %v, want = %v", file.data, []byte{})
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

func TestNewFile_WithInvalidName(t *testing.T) {
	id := "id"
	filename := "/filename"
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

func TestAddPermissions(t *testing.T) {
	id := "id"
	filename := "filename"

	file, err := NewFile(id, filename)
	if err != nil {
		t.Errorf("got error = %v, want = %v", err, nil)
		return
	}

	var uid int32 = 999
	file.AddPermissions(uid, Read)

	want := Read
	if got, exists := file.permissions[uid]; !exists || got != want {
		t.Errorf("got permissions = %v, want = %v", got, want)
	}

	file.AddPermissions(uid, Write)

	want = Write | Read
	if got, exists := file.permissions[uid]; !exists || got != want {
		t.Errorf("got permissions = %v, want = %v", got, want)
	}
}

func TestPermissions(t *testing.T) {
	id := "id"
	filename := "filename"

	file, err := NewFile(id, filename)
	if err != nil {
		t.Errorf("got error = %v, want = %v", err, nil)
		return
	}

	var uid int32 = 999
	var want uint8 = 0
	if got := file.Permissions(uid); got != want {
		t.Errorf("got permissions = %v, want = %v", got, want)
	}

	want = Write | Read
	file.AddPermissions(uid, want)

	if got := file.Permissions(uid); got != want {
		t.Errorf("got permissions = %v, want = %v", got, want)
	}
}

func TestRevokePermissions(t *testing.T) {
	id := "id"
	filename := "filename"

	file, err := NewFile(id, filename)
	if err != nil {
		t.Errorf("got error = %v, want = %v", err, nil)
		return
	}

	var uid int32 = 999
	file.AddPermissions(uid, Write|Read)
	file.RevokePermissions(uid, Write)

	want := Read
	if got, exists := file.permissions[uid]; !exists || got != want {
		t.Errorf("got permissions = %v, want = %v", got, want)
	}

	file.RevokePermissions(uid, Owner)
	if got, exists := file.permissions[uid]; !exists || got != want {
		t.Errorf("got permissions = %v, want = %v", got, want)
	}
}

func TestRevokeAccess(t *testing.T) {
	id := "id"
	filename := "filename"

	file, err := NewFile(id, filename)
	if err != nil {
		t.Errorf("got error = %v, want = %v", err, nil)
		return
	}

	var uid int32 = 999
	file.AddPermissions(uid, Write|Read)
	file.RevokeAccess(uid)

	if got, exists := file.permissions[uid]; exists {
		t.Errorf("got permissions = %v, want = %v", got, nil)
	}
}

func TestSharedWith(t *testing.T) {
	id := "id"
	filename := "filename"

	file, err := NewFile(id, filename)
	if err != nil {
		t.Errorf("got error = %v, want = %v", err, nil)
		return
	}

	want := []int32{777, 888, 999}
	for _, uid := range want {
		file.AddPermissions(uid, Read)
	}

	got := file.SharedWith()
	if len(got) != len(want) {
		t.Errorf("got shared with = %v, want = %v", got, want)
	}

	for _, wantUid := range want {
		found := false
		for _, uid := range got {
			if uid == wantUid {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("got shared with = %v, want = %v", got, want)
		}
	}
}

func TestOwners(t *testing.T) {
	id := "id"
	filename := "filename"

	file, err := NewFile(id, filename)
	if err != nil {
		t.Errorf("got error = %v, want = %v", err, nil)
		return
	}

	file.AddPermissions(777, Read)
	file.AddPermissions(888, Write|Owner)
	file.AddPermissions(999, Owner)

	want := []int32{888, 999}
	got := file.Owners()
	if len(got) != len(want) {
		t.Errorf("got shared with = %v, want = %v", got, want)
	}

	for _, wantUid := range want {
		found := false
		for _, uid := range got {
			if uid == wantUid {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("got shared with = %v, want = %v", got, want)
		}
	}
}

func TestAddValue(t *testing.T) {
	id := "id"
	filename := "filename"

	file, err := NewFile(id, filename)
	if err != nil {
		t.Errorf("got error = %v, want = %v", err, nil)
		return
	}

	key := "testing"
	want := "testing"
	file.AddValue(key, want)

	if got, exists := file.metadata[key]; !exists || got != want {
		t.Errorf("got metadata = %v, want = %v", got, want)
	}
}

func TestValue(t *testing.T) {
	id := "id"
	filename := "filename"

	file, err := NewFile(id, filename)
	if err != nil {
		t.Errorf("got error = %v, want = %v", err, nil)
		return
	}

	key := "testing"
	want := "testing"
	if got, exists := file.Value(key); exists {
		t.Errorf("got value = %v, want = %v", got, nil)
	}

	file.AddValue(key, want)
	if got, exists := file.Value(key); !exists || got != want {
		t.Errorf("got metadata = %v, want = %v", got, want)
	}
}
