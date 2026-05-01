package engine

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ignoredDirs are skipped during the repository walk; they contain generated
// artefacts or dependency caches that add no identification signal.
var ignoredDirs = map[string]struct{}{
	"node_modules": {},
	"vendor":       {},
	"__pycache__":  {},
	".venv":        {},
	"venv":         {},
	"env":          {},
	".dart_tool":   {},
	".gradle":      {},
	".next":        {},
	".nuxt":        {},
}

func scanRepository(rootPath string) (scanResult, error) {
	rootAbs, err := filepath.Abs(rootPath)
	if err != nil {
		return scanResult{}, err
	}

	res := scanResult{
		root:        rootAbs,
		projectName: filepath.Base(rootAbs),
		fileSet:     map[string]struct{}{},
		dirs:        map[string]struct{}{},
		extCounts:   map[string]int{},
		byBase:      map[string][]string{},
	}

	err = filepath.WalkDir(rootAbs, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		rel, err := filepath.Rel(rootAbs, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if rel == "." {
			return nil
		}

		if d.IsDir() {
			name := d.Name()
			if name == ".git" {
				return filepath.SkipDir
			}
			if _, skip := ignoredDirs[name]; skip {
				return filepath.SkipDir
			}
			res.dirs[rel] = struct{}{}
			return nil
		}

		res.files = append(res.files, rel)
		res.fileSet[rel] = struct{}{}

		base := filepath.Base(rel)
		res.byBase[base] = append(res.byBase[base], rel)
		ext := strings.ToLower(filepath.Ext(base))
		if ext != "" {
			res.extCounts[ext]++
		}
		return nil
	})
	if err != nil {
		return scanResult{}, err
	}

	sort.Strings(res.files)
	for k := range res.byBase {
		sort.Strings(res.byBase[k])
	}
	return res, nil
}

func (s scanResult) hasFile(path string) bool {
	_, ok := s.fileSet[path]
	return ok
}

func (s scanResult) hasBase(base string) bool {
	_, ok := s.byBase[base]
	return ok
}

func (s scanResult) pathsByBase(base string) []string {
	paths := s.byBase[base]
	out := make([]string, len(paths))
	copy(out, paths)
	return out
}

func (s scanResult) hasDir(path string) bool {
	_, ok := s.dirs[path]
	return ok
}

// hasBaseAtRoot returns true only when the given basename exists at the
// root of the repository (no path separators), preventing a nested file
// (e.g. android/Gemfile) from overriding a root-level marker.
func (s scanResult) hasBaseAtRoot(base string) bool {
	for _, path := range s.byBase[base] {
		if !strings.Contains(path, "/") {
			return true
		}
	}
	return false
}

func readFileIfExists(root, rel string) string {
	b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		return ""
	}
	return string(b)
}
