package directory

import (
	"context"
	"errors"
	"path"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	fb "github.com/alvidir/filebrowser"
	cert "github.com/alvidir/filebrowser/certificate"
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

func TestGetWhenDirectoryDoesNotExists(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	dirRepo := &directoryRepositoryMock{}
	dirRepo.findByUserId = func(ctx context.Context, userId int32) (*Directory, error) {
		return nil, fb.ErrNotFound
	}

	fileRepo := &fileRepositoryMock{}
	app := NewDirectoryApplication(dirRepo, fileRepo, logger)

	if _, err := app.Get(context.TODO(), 999, ""); !errors.Is(err, fb.ErrNotFound) {
		t.Errorf("got error = %v, want = %v", err, fb.ErrNotFound)
	}
}

func TestGet(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

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
	}

	for _, test := range tests {
		t := t
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			f, _ := file.NewFile("test", "filename")
			dirRepo := &directoryRepositoryMock{}
			files := make(map[string]*file.File)
			for _, fileName := range test.files {
				files[fileName] = f
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
			dir, err := app.Get(context.TODO(), 999, test.path)

			if err != nil {
				t.Errorf("got error = %v, want = %v", err, nil)
			} else if dir.id != "test" {
				t.Errorf("got id = %v, want = %v", dir.id, "test")
			} else if dir.userId != 999 {
				t.Errorf("got user id = %v, want = %v", dir.userId, 999)
			} else if testPath := filepath.Clean(test.path); dir.path != testPath {
				t.Errorf("got path = %v, want = %v", dir.path, testPath)
			} else if len(test.want) != len(dir.files) {
				t.Errorf("got files = %v, want = %v", dir.files, test.want)
			}

			for _, expectedPath := range test.want {
				if _, exists := dir.files[expectedPath]; !exists {
					t.Errorf("got files = %v, want = %v", dir.files, expectedPath)
				}
			}
		})
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

	if _, err := app.Delete(context.TODO(), 999, ""); !errors.Is(err, fb.ErrNotFound) {
		t.Errorf("got error = %v, want = %v", err, fb.ErrNotFound)
	}
}

func TestDeleteWhenUserIsSingleOwner(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	f, _ := file.NewFile("test", "filename")
	f.AddPermission(999, cert.Owner)

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
				"/path/to/file": f,
			},
		}, nil
	}

	fileRepo := &fileRepositoryMock{
		find: func(repo *fileRepositoryMock, ctx context.Context, id string) (*file.File, error) {
			return f, nil
		},

		delete: func(repo *fileRepositoryMock, ctx context.Context, file *file.File) error {
			return nil
		},
	}
	app := NewDirectoryApplication(dirRepo, fileRepo, logger)

	before := time.Now().Unix()
	if _, err := app.Delete(context.TODO(), 999, "/path/to/file"); err != nil {
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
	f.AddPermission(999, cert.Owner)
	f.AddPermission(888, cert.Owner)

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

	if _, err := app.Delete(context.TODO(), 999, ""); err != nil {
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

	if _, err := app.Delete(context.TODO(), 999, ""); err != nil {
		t.Errorf("got error = %v, want = %v", err, nil)
	}

	if deletedAt, exists := f.Value(file.MetadataDeletedAtKey); exists {
		t.Errorf("got deleted_at = %v, want = %v", deletedAt, nil)
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
	if _, err := app.RegisterFile(context.TODO(), f, 999, "path/to/file"); !errors.Is(err, fb.ErrNotFound) {
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
	if _, err := app.RegisterFile(context.TODO(), f, 999, "path/to/file"); err != nil {
		t.Errorf("got error = %v, want = %v", err, fb.ErrNotFound)
	}

	if got := d.files; len(got) != 1 {
		t.Errorf("got list len = %v, want = %v", len(got), 1)
	} else if got, exists := got["/path/to/file"]; !exists || got.Id() != "test" {
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
	f.AddPermission(999, cert.Read)
	d.AddFile(f, "path/to/file")

	if err := app.UnregisterFile(context.TODO(), f, 999); err != nil {
		t.Errorf("got error = %v, want = %v", err, fb.ErrNotFound)
	}

	if got, exists := d.files["path/to/file"]; exists {
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
	f.AddPermission(999, cert.Owner)
	d.AddFile(f, "path/to/file")

	if err := app.UnregisterFile(context.TODO(), f, 999); err != nil {
		t.Errorf("got error = %v, want = %v", err, fb.ErrNotFound)
	}

	if got, exists := d.files["path/to/file"]; exists {
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
	f.AddPermission(999, cert.Owner)
	f.AddPermission(888, cert.Read|cert.Write)

	d1.AddFile(f, "path/to/file")
	d2.AddFile(f, "path/to/file")

	if err := app.UnregisterFile(context.TODO(), f, 999); err != nil {
		t.Errorf("got error = %v, want = %v", err, fb.ErrNotFound)
	}

	if got, exists := d1.files["path/to/file"]; exists {
		t.Errorf("got file = %v, want = %v", got, nil)
	}

	if got, exists := d2.files["path/to/file"]; exists {
		t.Errorf("got file = %v, want = %v", got, nil)
	}

}

func TestMove(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	tests := []struct {
		name  string
		dest  string
		paths []string
		files []string
		want  []string
	}{
		{
			name:  "move to a new directory",
			dest:  "/new_directory/",
			paths: []string{"/a_file", "/another_file"},
			files: []string{
				"/a_file",
				"/another_file",
			},
			want: []string{
				"/new_directory/a_file",
				"/new_directory/another_file",
			},
		},
		{
			name:  "move to a existing directory",
			dest:  "/a_directory/",
			paths: []string{"/a_file"},
			files: []string{
				"/a_file",
				"/a_directory/another_file",
			},
			want: []string{
				"/a_directory/a_file",
				"/a_directory/another_file",
			},
		},
		{
			name:  "move directory to another directory",
			dest:  "/another_dir/a_directory/",
			paths: []string{"/a_directory/"},
			files: []string{
				"/a_directory/a_file",
				"/a_directory/another_file",
				"/another_dir/a_file",
			},
			want: []string{
				"/another_dir/a_directory/a_file",
				"/another_dir/a_directory/another_file",
				"/another_dir/a_file",
			},
		},
		{
			name:  "move file to parent directory",
			dest:  "/",
			paths: []string{"/a_directory/a_file"},
			files: []string{
				"/a_directory/a_file",
				"/a_directory/another_file",
			},
			want: []string{
				"/a_file",
				"/a_directory/another_file",
			},
		},
		{
			name:  "move to a directory with a file with the same name",
			dest:  "/a_directory/",
			paths: []string{"/a_file"},
			files: []string{
				"/a_file",
				"/a_directory/a_file",
			},
			want: []string{
				"/a_directory/a_file_1",
				"/a_directory/a_file",
			},
		},
		{
			name:  "move to new directory with same name as file",
			dest:  "/another_dir/an_item/a_directory/",
			paths: []string{"/a_directory/"},
			files: []string{
				"/a_directory/a_file",
				"/another_dir/an_item",
			},
			want: []string{
				"/another_dir/an_item_1/a_directory/a_file",
				"/another_dir/an_item",
			},
		},
		{
			name:  "move from a directory to another",
			dest:  "/a_directory/",
			paths: []string{"/another_dir/a_file"},
			files: []string{
				"/a_directory/a_file",
				"/another_dir/a_file",
			},
			want: []string{
				"/a_directory/a_file_1",
				"/a_directory/a_file",
			},
		},
		{
			name:  "move file with similar name to directory",
			dest:  "/a_directory/",
			paths: []string{"/a_file"},
			files: []string{
				"/a_file",
				"/a_file_1",
			},
			want: []string{
				"/a_directory/a_file",
				"/a_file_1",
			},
		},
		{
			name:  "move file with similar name to directory - 2",
			dest:  "/a_directory/",
			paths: []string{"/a_file_1"},
			files: []string{
				"/a_file",
				"/a_file_1",
			},
			want: []string{
				"/a_directory/a_file_1",
				"/a_file",
			},
		},
		{
			name:  "rename file",
			dest:  "/renamed_file",
			paths: []string{"/a_file"},
			files: []string{
				"/a_file",
				"/a_file_1",
			},
			want: []string{
				"/renamed_file",
				"/a_file_1",
			},
		},
		{
			name:  "rename deeper file",
			dest:  "/a_directory/renamed_file",
			paths: []string{"/a_directory/a_file"},
			files: []string{
				"/a_file",
				"/a_directory/a_file",
				"/a_directory/another_file",
			},
			want: []string{
				"/a_file",
				"/a_directory/renamed_file",
				"/a_directory/another_file",
			},
		},
		{
			name:  "rename directory",
			dest:  "/renamed_dir/",
			paths: []string{"/a_directory/"},
			files: []string{
				"/a_file",
				"/a_directory/a_file",
				"/a_directory/another_file",
			},
			want: []string{
				"/a_file",
				"/renamed_dir/a_file",
				"/renamed_dir/another_file",
			},
		},
		{
			name:  "rename deeper directory",
			dest:  "/a_directory/renamed_dir/",
			paths: []string{"/a_directory/another_dir/"},
			files: []string{
				"/a_file",
				"/a_directory/a_file",
				"/a_directory/another_file",
				"/a_directory/another_dir/a_file",
				"/a_directory/another_dir/another_file",
			},
			want: []string{
				"/a_file",
				"/a_directory/a_file",
				"/a_directory/another_file",
				"/a_directory/renamed_dir/a_file",
				"/a_directory/renamed_dir/another_file",
			},
		},
		{
			name:  "move directory where file with the same name",
			dest:  "/another_dir/a_directory",
			paths: []string{"/a_directory"},
			files: []string{
				"/a_file",
				"/a_directory/a_file",
				"/a_directory/another_file",
				"/another_dir/a_directory",
				"/another_dir/nested_dir/another_file",
			},
			want: []string{
				"/a_file",
				"/another_dir/a_directory_1/a_file",
				"/another_dir/a_directory_1/another_file",
				"/another_dir/a_directory",
				"/another_dir/nested_dir/another_file",
			},
		},
	}

	for _, test := range tests {
		t := t
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			dirRepo := &directoryRepositoryMock{}
			files := make(map[string]*file.File)
			for id, fileName := range test.files {
				files[fileName], _ = file.NewFile(strconv.Itoa(id), "filename")
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
					for _, f := range files {
						if f.Id() == id {
							return f, nil
						}
					}

					return nil, nil
				},
			}

			app := NewDirectoryApplication(dirRepo, fileRepo, logger)

			_, err := app.Move(context.TODO(), 999, test.paths, test.dest)
			if err != nil {
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

func TestDeleteFiles(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	tests := []struct {
		name  string
		path  string
		files []string
		want  []string
	}{
		{
			name: "filter single file",
			path: "/a_file",
			files: []string{
				"/a_file",
				"/a_directory/a_file",
			},
			want: []string{
				"/a_directory/a_file",
			},
		},
		{
			name: "filter directory",
			path: "/a_directory/",
			files: []string{
				"/a_file",
				"/a_directory/a_file",
				"/a_directory/another_file",
			},
			want: []string{
				"/a_file",
			},
		},
	}

	for _, test := range tests {
		t := t
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			dirRepo := &directoryRepositoryMock{}
			files := make(map[string]*file.File)
			for index, fp := range test.files {
				id := strconv.Itoa(index)
				files[fp], _ = file.NewFile(id, path.Base(fp))
				files[fp].SetDirectory(path.Dir(fp))
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
					for _, f := range files {
						if f.Id() == id {
							return f, nil
						}
					}

					return nil, fb.ErrNotFound
				},
			}

			app := NewDirectoryApplication(dirRepo, fileRepo, logger)

			_, err := app.Delete(context.TODO(), 999, test.path)
			if err != nil {
				t.Errorf("got error = %v, want = nil", err)
			}

			filenames := make([]string, 0, len(files))
			for _, f := range files {
				fp := filepath.Join(PathSeparator, f.Directory(), f.Name())
				filenames = append(filenames, fp)
			}

			if len(test.want) != len(files) {
				t.Errorf("got files = %+v, want = %v", filenames, test.want)
			}

			for _, expectedPath := range test.want {
				if _, exists := files[expectedPath]; !exists {
					t.Errorf("got files = %v, want = %v", filenames, test.want)
				}
			}

		})
	}
}
