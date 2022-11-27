package directory

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	fb "github.com/alvidir/filebrowser"
	"github.com/alvidir/filebrowser/file"
	"go.uber.org/zap"
)

const (
	mockDirectoryId = "000"
)

type directoryRepositoryMock struct {
	findByUserId func(ctx context.Context, userId int32) (*Directory, error)
	create       func(ctx context.Context, dir *Directory) error
	save         func(ctx context.Context, dir *Directory) error
	delete       func(ctx context.Context, dir *Directory) error
}

func (mock *directoryRepositoryMock) FindByUserId(ctx context.Context, userId int32) (*Directory, error) {
	if mock.findByUserId != nil {
		return mock.findByUserId(ctx, userId)
	}

	dir := NewDirectory(userId)
	dir.id = mockDirectoryId
	return dir, nil
}

func (mock *directoryRepositoryMock) Create(ctx context.Context, dir *Directory) error {
	if mock.create != nil {
		return mock.create(ctx, dir)
	}

	dir.id = mockDirectoryId
	return nil
}

func (mock *directoryRepositoryMock) Save(ctx context.Context, dir *Directory) error {
	if mock.save != nil {
		return mock.save(ctx, dir)
	}

	dir.id = mockDirectoryId
	return nil
}

func (mock *directoryRepositoryMock) Delete(ctx context.Context, dir *Directory) error {
	if mock.delete != nil {
		return mock.delete(ctx, dir)
	}

	dir.id = mockDirectoryId
	return nil
}

type fileRepositoryMock struct {
	create  func(repo *fileRepositoryMock, ctx context.Context, file *file.File) error
	find    func(repo *fileRepositoryMock, ctx context.Context, id string) (*file.File, error)
	findAll func(repo *fileRepositoryMock, ctx context.Context, ids []string) ([]*file.File, error)
	save    func(repo *fileRepositoryMock, ctx context.Context, file *file.File) error
	delete  func(repo *fileRepositoryMock, ctx context.Context, file *file.File) error
}

func (mock *fileRepositoryMock) Create(ctx context.Context, file *file.File) error {
	if mock.create != nil {
		return mock.create(mock, ctx, file)
	}

	return fb.ErrNotFound
}

func (mock *fileRepositoryMock) Find(ctx context.Context, id string) (*file.File, error) {
	if mock.find != nil {
		return mock.find(mock, ctx, id)
	}

	return nil, fb.ErrNotFound
}

func (mock *fileRepositoryMock) FindAll(ctx context.Context, ids []string) ([]*file.File, error) {
	if mock.findAll != nil {
		return mock.findAll(mock, ctx, ids)
	}

	return nil, fb.ErrNotFound
}

func (mock *fileRepositoryMock) Save(ctx context.Context, file *file.File) error {
	if mock.save != nil {
		return mock.save(mock, ctx, file)
	}

	return fb.ErrUnknown
}

func (mock *fileRepositoryMock) Delete(ctx context.Context, file *file.File) error {
	if mock.delete != nil {
		return mock.delete(mock, ctx, file)
	}

	return fb.ErrUnknown
}

func TestCreateWhenAlreadyExists(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	dirRepo := &directoryRepositoryMock{}
	dirRepo.findByUserId = func(ctx context.Context, userId int32) (*Directory, error) {
		return nil, nil
	}

	fileRepo := &fileRepositoryMock{}
	app := NewDirectoryApplication(dirRepo, fileRepo, logger)

	var want int32 = 999
	if _, err := app.Create(context.TODO(), want); !errors.Is(err, fb.ErrAlreadyExists) {
		t.Errorf("got error = %v, want = %v", err, fb.ErrAlreadyExists)
	}
}

func TestCreate(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	dirRepo := &directoryRepositoryMock{}
	dirRepo.findByUserId = func(ctx context.Context, userId int32) (*Directory, error) {
		return nil, fb.ErrNotFound
	}

	fileRepo := &fileRepositoryMock{}
	app := NewDirectoryApplication(dirRepo, fileRepo, logger)

	var want int32 = 999
	dir, err := app.Create(context.TODO(), want)
	if err != nil {
		t.Errorf("got error = %v, want = %v", err, nil)
	}

	if got := dir.id; got != mockDirectoryId {
		t.Errorf("got directory.id = %v, want = %v", got, mockDirectoryId)
	}

	if got := dir.userId; got != want {
		t.Errorf("got directory.userId = %v, want = %v", got, want)
	}
}

func TestRetrieveWhenDirectoryDoesNotExists(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	dirRepo := &directoryRepositoryMock{}
	dirRepo.findByUserId = func(ctx context.Context, userId int32) (*Directory, error) {
		return nil, fb.ErrNotFound
	}

	fileRepo := &fileRepositoryMock{}
	app := NewDirectoryApplication(dirRepo, fileRepo, logger)

	if _, err := app.Retrieve(context.TODO(), 999, "", ""); !errors.Is(err, fb.ErrNotFound) {
		t.Errorf("got error = %v, want = %v", err, fb.ErrNotFound)
	}
}

func TestRetrieve(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	dirRepo := &directoryRepositoryMock{}
	dirRepo.findByUserId = func(ctx context.Context, userId int32) (*Directory, error) {
		f, _ := file.NewFile("", "filename")
		return &Directory{
			id:     "test",
			userId: 999,
			files: map[string]*file.File{
				"a_file": f,
			},
		}, nil
	}

	fileRepo := &fileRepositoryMock{}
	fileRepo.findAll = func(repo *fileRepositoryMock, ctx context.Context, ids []string) ([]*file.File, error) {
		return nil, nil
	}

	app := NewDirectoryApplication(dirRepo, fileRepo, logger)

	if dir, err := app.Retrieve(context.TODO(), 999, "", ""); err != nil {
		t.Errorf("got error = %v, want = %v", err, nil)
	} else if dir.id != "test" {
		t.Errorf("got id = %v, want = %v", dir.id, "test")
	} else if dir.userId != 999 {
		t.Errorf("got user id = %v, want = %v", dir.userId, 999)
	}
}

func TestDeleteWhenDirectoryDoesNotExists(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	dirRepo := &directoryRepositoryMock{}
	dirRepo.findByUserId = func(ctx context.Context, userId int32) (*Directory, error) {
		return nil, fb.ErrNotFound
	}

	fileRepo := &fileRepositoryMock{}
	app := NewDirectoryApplication(dirRepo, fileRepo, logger)

	if err := app.Delete(context.TODO(), 999); !errors.Is(err, fb.ErrNotFound) {
		t.Errorf("got error = %v, want = %v", err, fb.ErrNotFound)
	}
}

func TestDeleteWhenUserIsSingleOwner(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	f, _ := file.NewFile("test", "filename")
	f.AddPermission(999, fb.Owner)

	dirRepo := &directoryRepositoryMock{
		delete: func(ctx context.Context, dir *Directory) error {
			return nil
		},
	}

	dirRepo.findByUserId = func(ctx context.Context, userId int32) (*Directory, error) {
		return &Directory{
			id:     "test",
			userId: 999,
			files: map[string]*file.File{
				"path/to/file": f,
			},
		}, nil
	}

	fileRepo := &fileRepositoryMock{
		find: func(repo *fileRepositoryMock, ctx context.Context, id string) (*file.File, error) {
			return f, nil
		},
	}
	app := NewDirectoryApplication(dirRepo, fileRepo, logger)

	before := time.Now().Unix()
	if err := app.Delete(context.TODO(), 999); err != nil {
		t.Errorf("got error = %v, want = %v", err, nil)
	}

	after := time.Now().Unix()

	if deletedAt, exists := f.Value(file.MetadataDeletedAtKey); !exists {
		t.Errorf("got deleted_at = %v, want > %v && < %v", deletedAt, before, after)
	} else if unixDeletedAt, err := strconv.ParseInt(deletedAt, file.TimestampBase, 64); err != nil {
		t.Errorf("got error = %v, want = %v", err, nil)
	} else if unixDeletedAt > before || unixDeletedAt < after {
		t.Errorf("got deleted_at = %v, want > %v && < %v", unixDeletedAt, before, after)
	}
}

func TestDeleteWhenUserIsNotSingleOwner(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	f, _ := file.NewFile("test", "filename")
	f.AddPermission(999, fb.Owner)
	f.AddPermission(888, fb.Owner)

	dirRepo := &directoryRepositoryMock{
		delete: func(ctx context.Context, dir *Directory) error {
			return nil
		},
	}

	dirRepo.findByUserId = func(ctx context.Context, userId int32) (*Directory, error) {
		return &Directory{
			id:     "test",
			userId: 999,
			files: map[string]*file.File{
				"path/to/file": f,
			},
		}, nil
	}

	fileRepo := &fileRepositoryMock{
		find: func(repo *fileRepositoryMock, ctx context.Context, id string) (*file.File, error) {
			return f, nil
		},
	}
	app := NewDirectoryApplication(dirRepo, fileRepo, logger)

	if err := app.Delete(context.TODO(), 999); err != nil {
		t.Errorf("got error = %v, want = %v", err, nil)
	}

	if deletedAt, exists := f.Value(file.MetadataDeletedAtKey); exists {
		t.Errorf("got deleted_at = %v, want = %v", deletedAt, nil)
	}
}

func TestDeleteWhenUserIsNotOwner(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	f, _ := file.NewFile("test", "filename")
	dirRepo := &directoryRepositoryMock{
		delete: func(ctx context.Context, dir *Directory) error {
			return nil
		},
	}

	dirRepo.findByUserId = func(ctx context.Context, userId int32) (*Directory, error) {
		return &Directory{
			id:     "test",
			userId: 999,
			files: map[string]*file.File{
				"path/to/file": f,
			},
		}, nil
	}

	fileRepo := &fileRepositoryMock{
		find: func(repo *fileRepositoryMock, ctx context.Context, id string) (*file.File, error) {
			return f, nil
		},
	}
	app := NewDirectoryApplication(dirRepo, fileRepo, logger)

	if err := app.Delete(context.TODO(), 999); err != nil {
		t.Errorf("got error = %v, want = %v", err, nil)
	}

	if deletedAt, exists := f.Value(file.MetadataDeletedAtKey); exists {
		t.Errorf("got deleted_at = %v, want = %v", deletedAt, nil)
	}
}

func TestRelocate_2(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	f, _ := file.NewFile("test", "filename")
	dirRepo := &directoryRepositoryMock{}
	files := map[string]*file.File{
		"folder/another folder 1/another folder/world.txt": f,
	}

	dirRepo.findByUserId = func(ctx context.Context, userId int32) (*Directory, error) {
		return &Directory{
			id:     "test",
			userId: 999,
			files:  files,
		}, nil
	}

	fileRepo := &fileRepositoryMock{
		find: func(repo *fileRepositoryMock, ctx context.Context, id string) (*file.File, error) {
			return f, nil
		},
	}

	app := NewDirectoryApplication(dirRepo, fileRepo, logger)

	err := app.Relocate(context.TODO(), 999, "root", "^/?folder")
	if err != nil {
		t.Errorf("got error = %v, want = %v", err, nil)
	}

	want := "root/folder/another folder 1/another folder/world.txt"
	if _, exists := files[want]; !exists {
		t.Errorf("got files = %v, want = %s", files, want)
	}
}

func TestRelocate_3(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	f, _ := file.NewFile("test", "filename")
	dirRepo := &directoryRepositoryMock{}
	files := map[string]*file.File{
		"folder/another folder 1/another folder/world.txt": f,
	}

	dirRepo.findByUserId = func(ctx context.Context, userId int32) (*Directory, error) {
		return &Directory{
			id:     "test",
			userId: 999,
			files:  files,
		}, nil
	}

	fileRepo := &fileRepositoryMock{
		find: func(repo *fileRepositoryMock, ctx context.Context, id string) (*file.File, error) {
			return f, nil
		},
	}

	app := NewDirectoryApplication(dirRepo, fileRepo, logger)

	err := app.Relocate(context.TODO(), 999, "folder1", "^/?folder")
	if err != nil {
		t.Errorf("got error = %v, want = %v", err, nil)
	}

	want := "folder1/folder/another folder 1/another folder/world.txt"
	if _, exists := files[want]; !exists {
		t.Errorf("got files = %v, want = %s", files, want)
	}
}

func TestRelocate(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	tests := []struct {
		name   string
		target string
		filter string
		want   []string
		err    error
	}{
		{
			name:   "move to a new directory",
			target: "new_directory",
			filter: "^/?((a_file)|(another_file))",
			want: []string{
				"new_directory/a_file",
				"new_directory/another_file",
				"a_directory/a_file",
				"/a_directory/another_file",
				"another_dir/test.txt",
				"unique_name",
			},
			err: nil,
		},
		{
			name:   "move to a existing directory",
			target: "a_directory",
			filter: "^/?unique_name",
			want: []string{
				"a_file",
				"/another_file",
				"a_directory/a_file",
				"/a_directory/another_file",
				"a_directory/unique_name",
				"another_dir/test.txt",
			},
			err: nil,
		},
		{
			name:   "move directory",
			target: "another_dir",
			filter: "^/?a_directory",
			want: []string{
				"a_file",
				"/another_file",
				"another_dir/a_directory/a_file",
				"another_dir/a_directory/another_file",
				"another_dir/test.txt",
				"unique_name",
			},
			err: nil,
		},
		{
			name:   "move to a directory with a file with the same name",
			target: "a_directory",
			filter: "^/?a_file",
			want: []string{
				"a_file",
				"/another_file",
				"a_directory/a_file",
				"/a_directory/another_file",
				"another_dir/test.txt",
				"unique_name",
			},
			err: fb.ErrAlreadyExists,
		},
		{
			name:   "move directory with already existing name",
			target: "another_dir/test.txt",
			filter: "^/?a_directory",
			want: []string{
				"a_file",
				"/another_file",
				"a_directory/a_file",
				"/a_directory/another_file",
				"another_dir/test.txt",
				"unique_name",
			},
			err: fb.ErrAlreadyExists,
		},
	}

	for _, test := range tests {
		t := t
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			f, _ := file.NewFile("test", "filename")
			dirRepo := &directoryRepositoryMock{}
			files := map[string]*file.File{
				"a_file":                    f,
				"/another_file":             f,
				"a_directory/a_file":        f,
				"/a_directory/another_file": f,
				"another_dir/test.txt":      f,
				"unique_name":               f,
			}

			dirRepo.findByUserId = func(ctx context.Context, userId int32) (*Directory, error) {
				return &Directory{
					id:     "test",
					userId: 999,
					files:  files,
				}, nil
			}

			fileRepo := &fileRepositoryMock{
				find: func(repo *fileRepositoryMock, ctx context.Context, id string) (*file.File, error) {
					return f, nil
				},
			}

			app := NewDirectoryApplication(dirRepo, fileRepo, logger)

			err := app.Relocate(context.TODO(), 999, test.target, test.filter)
			if test.err != nil && !errors.Is(err, test.err) {
				t.Errorf("got error = %v, want = %v", err, test.err)
			}

			if test.err == nil && err != nil {
				t.Errorf("got error = %v, want = nil", err)
			}

			if len(test.want) != len(files) {
				t.Errorf("got files = %v, want = %v", files, test.want)
			}

			for _, expectedPath := range test.want {
				if _, exists := files[expectedPath]; !exists {
					t.Errorf("got files = %v, want = %v", files, test.want)
				}
			}

		})
	}
}

func TestRegisterFileWhenDirectoryDoesNotExists(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	dirRepo := &directoryRepositoryMock{}
	dirRepo.findByUserId = func(ctx context.Context, userId int32) (*Directory, error) {
		return nil, fb.ErrNotFound
	}

	fileRepo := &fileRepositoryMock{}
	app := NewDirectoryApplication(dirRepo, fileRepo, logger)

	f, _ := file.NewFile("test", "filename")
	if err := app.RegisterFile(context.TODO(), f, 999, "path/to/file"); !errors.Is(err, fb.ErrNotFound) {
		t.Errorf("got error = %v, want = %v", err, fb.ErrNotFound)
	}
}

func TestRegisterFile(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	d := NewDirectory(999)

	dirRepo := &directoryRepositoryMock{}
	dirRepo.findByUserId = func(ctx context.Context, userId int32) (*Directory, error) {
		return d, nil
	}

	fileRepo := &fileRepositoryMock{}
	app := NewDirectoryApplication(dirRepo, fileRepo, logger)

	f, _ := file.NewFile("test", "filename")
	if err := app.RegisterFile(context.TODO(), f, 999, "path/to/file"); err != nil {
		t.Errorf("got error = %v, want = %v", err, fb.ErrNotFound)
	}

	if got := d.Files(); len(got) != 1 {
		t.Errorf("got list len = %v, want = %v", len(got), 1)
	} else if got, exists := got["path/to/file"]; !exists || got.Id() != "test" {
		t.Errorf("got file = %v, want = %v", got, "test")
	}
}

func TestUnregisterFileWhenDirectoryDoesNotExists(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	dirRepo := &directoryRepositoryMock{}
	dirRepo.findByUserId = func(ctx context.Context, userId int32) (*Directory, error) {
		return nil, fb.ErrNotFound
	}

	fileRepo := &fileRepositoryMock{}
	app := NewDirectoryApplication(dirRepo, fileRepo, logger)

	f, _ := file.NewFile("test", "filename")
	if err := app.UnregisterFile(context.TODO(), f, 999); !errors.Is(err, fb.ErrNotFound) {
		t.Errorf("got error = %v, want = %v", err, fb.ErrNotFound)
	}
}

func TestUnregisterFileWhenUserIsNoOwner(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	d := NewDirectory(999)

	dirRepo := &directoryRepositoryMock{}
	dirRepo.findByUserId = func(ctx context.Context, userId int32) (*Directory, error) {
		return d, nil
	}

	fileRepo := &fileRepositoryMock{}
	app := NewDirectoryApplication(dirRepo, fileRepo, logger)

	f, _ := file.NewFile("test", "filename")
	f.AddPermission(999, fb.Read)
	d.AddFile(f, "path/to/file")

	if err := app.UnregisterFile(context.TODO(), f, 999); err != nil {
		t.Errorf("got error = %v, want = %v", err, fb.ErrNotFound)
	}

	if got, exists := d.Files()["path/to/file"]; exists {
		t.Errorf("got file = %v, want = %v", got, exists)
	}
}

func TestUnregisterFileWhenFileIsDeleted(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	d := NewDirectory(999)

	dirRepo := &directoryRepositoryMock{}
	dirRepo.findByUserId = func(ctx context.Context, userId int32) (*Directory, error) {
		return d, nil
	}

	fileRepo := &fileRepositoryMock{}
	app := NewDirectoryApplication(dirRepo, fileRepo, logger)

	f, _ := file.NewFile("test", "filename")
	f.AddMetadata(file.MetadataDeletedAtKey, strconv.FormatInt(time.Now().Unix(), file.TimestampBase))
	f.AddPermission(999, fb.Owner)
	d.AddFile(f, "path/to/file")

	if err := app.UnregisterFile(context.TODO(), f, 999); err != nil {
		t.Errorf("got error = %v, want = %v", err, fb.ErrNotFound)
	}

	if got, exists := d.Files()["path/to/file"]; exists {
		t.Errorf("got file = %v, want = %v", got, nil)
	}
}

func TestUnregisterFileWhenFileIsShared(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	d1 := NewDirectory(999)
	d2 := NewDirectory(888)

	dirRepo := &directoryRepositoryMock{}
	dirRepo.findByUserId = func(ctx context.Context, userId int32) (*Directory, error) {
		if userId == 999 {
			return d1, nil
		}

		if userId == 888 {
			return d2, nil
		}

		return nil, fb.ErrNotFound
	}

	fileRepo := &fileRepositoryMock{}
	app := NewDirectoryApplication(dirRepo, fileRepo, logger)

	f, _ := file.NewFile("test", "filename")
	f.AddMetadata(file.MetadataDeletedAtKey, strconv.FormatInt(time.Now().Unix(), file.TimestampBase))
	f.AddPermission(999, fb.Owner)
	f.AddPermission(888, fb.Read|fb.Write)

	d1.AddFile(f, "path/to/file")
	d2.AddFile(f, "path/to/file")

	if err := app.UnregisterFile(context.TODO(), f, 999); err != nil {
		t.Errorf("got error = %v, want = %v", err, fb.ErrNotFound)
	}

	if got, exists := d1.Files()["path/to/file"]; exists {
		t.Errorf("got file = %v, want = %v", got, nil)
	}

	if got, exists := d2.Files()["path/to/file"]; exists {
		t.Errorf("got file = %v, want = %v", got, nil)
	}

}

func TestFilterByDir(t *testing.T) {
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
			filterFn, err := NewFilterByDirFn(test.path)
			if err != nil {
				t.Errorf("got error = %v, want = nil", err)
			}

			files, err := subject.FilterFiles([]FilterFileFn{filterFn})
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
