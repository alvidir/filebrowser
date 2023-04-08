package directory

import (
	"fmt"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/alvidir/filebrowser/file"
)

const (
	PathSeparator = "/"
)

type SearchMatch struct {
	file  *file.File
	start int
	end   int
}

type Directory struct {
	id     string
	userId int32
	files  map[string]*file.File
	path   string
}

func NewDirectory(userId int32) *Directory {
	return &Directory{
		id:     "",
		userId: userId,
		files:  make(map[string]*file.File),
	}
}

func pathComponents(p string) []string {
	paths := strings.Split(p, PathSeparator)

	components := make([]string, 0, len(paths))
	components = append(components, PathSeparator)

	for _, p := range paths {
		if len(p) > 0 {
			components = append(components, p)
		}
	}

	return components
}

func (dir *Directory) getAvailablePath(dest string) string {
	components := pathComponents(dest)
	for index := 0; index < len(components); index++ {
		candidate := components[index]
		counter := 1

		for {
			subject := filepath.Join(components[0 : index+1]...)
			if _, exists := dir.files[subject]; !exists {
				break
			}

			components[index] = fmt.Sprintf("%s_%d", candidate, counter)
			counter++
		}
	}

	return filepath.Join(components...)
}

func (dir *Directory) AddFile(file *file.File, fp string) string {
	fp = dir.getAvailablePath(fp)
	file.SetDirectory(path.Dir(fp))
	dir.files[fp] = file
	return fp
}

func (dir *Directory) RemoveFile(file *file.File) {
	for fp, f := range dir.files {
		if f.Id() == file.Id() {
			delete(dir.files, fp)
			return
		}
	}
}

func (dir *Directory) FilesByPath(p string) map[string]*file.File {
	absP := filepath.Join(PathSeparator, p)
	if absP == PathSeparator {
		return dir.files
	}

	files := make(map[string]*file.File)
	for fp, f := range dir.files {
		absFp := filepath.Join(PathSeparator, fp)

		if strings.HasPrefix(absFp, absP) &&
			(len(absP) == len(absFp) || path.IsAbs(absFp[len(absP):])) {
			files[absFp] = f
		}

	}

	return files
}

func (dir *Directory) FileByPath(p string) *file.File {
	absP := filepath.Join(PathSeparator, p)
	return dir.files[absP]
}

func (dir *Directory) AggregateFiles(p string) map[string]*file.File {
	files := make(map[string]*file.File)

	type FolderAggregate struct {
		size      int
		updatedAt int
	}

	folders := make(map[string]*FolderAggregate)

	absP := filepath.Join(PathSeparator, p)
	pCount := strings.Count(p, PathSeparator)

	if absP != PathSeparator {
		pCount++
	}

	for absFp, f := range dir.FilesByPath(p) {
		if pCount < strings.Count(absFp, PathSeparator) {
			updatedAt := 0
			if sizeStr, exists := f.Metadata()[file.MetadataUpdatedAtKey]; exists {
				if unix, err := strconv.ParseInt(sizeStr, file.TimestampBase, 64); err == nil {
					updatedAt = int(unix)
				}
			}

			// f is located deeper in the directory tree, and so, there is a folder at absP containing it
			folderPath := filepath.Join(pathComponents(absFp)[0 : pCount+1]...)
			if folder, exists := folders[folderPath]; exists {
				folder.size += 1

				if updatedAt > folder.updatedAt {
					folder.updatedAt = updatedAt
				}
			} else {
				folders[folderPath] = &FolderAggregate{
					size:      1,
					updatedAt: updatedAt,
				}
			}

			continue
		}

		f.MarkAsProtected() // avoid saving changes
		f.SetDirectory(absP)
		f.SetName(path.Base(absFp))
		files[absFp] = f
	}

	for folderPath, aggregate := range folders {
		folder, err := file.NewFile("", path.Base(folderPath))
		if err != nil {
			continue
		}

		folder.SetFlag(file.Directory)
		folder.SetDirectory(path.Dir(folderPath))
		folder.AddMetadata(file.MetadataSizeKey, strconv.Itoa(aggregate.size))
		folder.AddMetadata(file.MetadataUpdatedAtKey, strconv.FormatInt(int64(aggregate.updatedAt), file.TimestampBase))
		files[folderPath] = folder
	}

	return files
}

func allSubstringOcurrences(s string, substr string) [][]int {
	ocurrences := make([][]int, 0)
	offset := 0

	for len(s) >= len(substr) {
		index := strings.Index(s, substr)
		if index == -1 {
			break
		}

		start := index + offset
		ocurrences = append(ocurrences, []int{start, start + len(substr)})
		offset += index + len(substr)
		s = s[index+len(substr):]
	}

	return ocurrences
}

func (dir *Directory) Search(search string) []SearchMatch {
	searchMatches := make([]SearchMatch, 0)
	matchingfiles := make(map[string]*file.File)

	for fp, f := range dir.files {
		absFp := filepath.Join(PathSeparator, fp)

		offset := 0
		if search == PathSeparator {
			offset = 1
		}

		for _, match := range allSubstringOcurrences(strings.ToLower(absFp), search) {
			matchingFile := f

			if end := match[1]; end+offset <= len(path.Dir(absFp)) {
				// the match occurs somewhere in the file's directory
				directory := path.Dir(absFp[:end])
				paths := pathComponents(absFp[len(directory):])
				name := paths[1] // 1 since pathComponents includes "/" at the beginning

				matchingFile, _ = file.NewFile("", name)
				matchingFile.SetDirectory(directory)
				matchingFile.SetFlag(file.Directory)
			} else {
				// the match occurs somewhere in the filepath
				matchingFile.MarkAsProtected()
				matchingFile.SetDirectory(path.Dir(absFp))
				matchingFile.SetName(path.Base(absFp))
			}

			filepath := filepath.Join(matchingFile.Directory(), matchingFile.Name())
			if _, exists := matchingfiles[filepath]; exists {
				continue
			}

			searchMatches = append(searchMatches, SearchMatch{
				file:  matchingFile,
				start: match[0],
				end:   match[1],
			})

			matchingfiles[filepath] = matchingFile
		}
	}

	return searchMatches
}
