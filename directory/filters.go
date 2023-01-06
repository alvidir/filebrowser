package directory

import (
	"regexp"
	"strconv"
	"strings"

	fb "github.com/alvidir/filebrowser"
	"github.com/alvidir/filebrowser/file"
)

type filterFileFn func(string, *file.File) (string, *file.File)
type filterFileItems map[*file.File]string

func newFilterByRegexFn(target string) (filterFileFn, error) {
	regex, err := regexp.Compile(target)
	if err != nil {
		return nil, fb.ErrInvalidFormat
	}

	filterFn := func(p string, f *file.File) (string, *file.File) {
		if regex.MatchString(p) {
			return p, f
		}

		return "", nil
	}

	return filterFn, nil
}

func newFilterByPrefixFn(target string) (filterFileFn, error) {
	target = fb.NormalizePath(target)
	filterFn := func(p string, f *file.File) (string, *file.File) {
		p = fb.NormalizePath(p)
		if strings.HasPrefix(p, target) {
			return p, f
		}

		return "", nil
	}

	return filterFn, nil
}

func newFilterByDirFn(target string) (filterFileFn, error) {
	target = fb.NormalizePath(target)

	depth := len(fb.PathComponents(target))
	filterFn := func(p string, f *file.File) (string, *file.File) {
		p = fb.NormalizePath(p)

		if strings.Compare(p, target) == 0 {
			// 0 means p == target, so is not filtering by path, but by a filename
			return "", nil
		} else if !strings.HasPrefix(p, target) {
			return "", nil
		}

		items := fb.PathComponents(p)
		name := items[depth]

		if len(items) > depth+1 {
			f.SetFlag(file.Directory | file.Blurred)
			f.SetName(name)
		}

		return name, f
	}

	return filterFn, nil
}

func filterFiles(files map[string]*file.File, filters []filterFileFn) (filterFileItems, error) {
	filtered := make(filterFileItems)

	for p, f := range files {
		selected := f

		for _, filter := range filters {
			if filter == nil {
				continue
			}

			if p, selected = filter(p, f); selected == nil {
				break
			}
		}

		if selected != nil {
			filtered[selected] = p
		}
	}

	return filtered, nil
}

func (files filterFileItems) Range(fn func(string, *file.File) bool) {
	for f, p := range files {
		if !fn(p, f) {
			break
		}
	}
}

func (files filterFileItems) Agregate() (map[string]*file.File, error) {
	transposed := make(map[string]*file.File)

	type agregation struct {
		createdAt int64
		updatedAt int64
		size      int64
	}

	agregations := make(map[string]*agregation)

	for f, p := range files {
		transposed[p] = f
		if f.Flags()&file.Directory != 0 {
			if _, exists := agregations[p]; !exists {
				agregations[p] = &agregation{createdAt: -1, updatedAt: -1, size: 0}
			}

			agregation := agregations[p]
			createdAt, err := strconv.ParseInt(f.Metadata()[file.MetadataCreatedAtKey], file.TimestampBase, 64)
			if err != nil {
				return nil, err
			}

			if agregation.createdAt < 0 || agregation.createdAt > createdAt {
				agregation.createdAt = createdAt
			}

			updatedAt, err := strconv.ParseInt(f.Metadata()[file.MetadataUpdatedAtKey], file.TimestampBase, 64)
			if err != nil {
				return nil, err
			}

			if agregation.updatedAt < 0 || agregation.updatedAt < updatedAt {
				agregation.updatedAt = updatedAt
			}

			agregation.size++
		}
	}

	for key, agregation := range agregations {
		transposed[key].AddMetadata(file.MetadataCreatedAtKey, strconv.FormatInt(agregation.createdAt, file.TimestampBase))
		transposed[key].AddMetadata(file.MetadataUpdatedAtKey, strconv.FormatInt(agregation.updatedAt, file.TimestampBase))
		transposed[key].AddMetadata(file.MetadataSizeKey, strconv.FormatInt(agregation.size, file.TimestampBase))
		transposed[key].SetID("")
	}

	return transposed, nil
}
