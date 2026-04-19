package engine

import (
	"encoding/json"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var extToLanguage = map[string]string{
	".go":    "Go",
	".js":    "JavaScript",
	".mjs":   "JavaScript",
	".cjs":   "JavaScript",
	".ts":    "TypeScript",
	".tsx":   "TypeScript",
	".py":    "Python",
	".rs":    "Rust",
	".java":  "Java",
	".kt":    "Kotlin",
	".kts":   "Kotlin",
	".dart":  "Dart",
	".php":   "PHP",
	".rb":    "Ruby",
	".cs":    "C#",
	".cpp":   "C++",
	".cc":    "C++",
	".cxx":   "C++",
	".hpp":   "C++",
	".h":     "C",
	".swift": "Swift",
	".scala": "Scala",
	".lua":   "Lua",
	".sh":    "Shell",
	".bash":  "Shell",
	".html":  "FrontendAssets",
	".css":   "FrontendAssets",
	".scss":  "FrontendAssets",
}

type detected struct {
	projectType      string
	languages        []string
	frameworks       []string
	runtime          string
	packageManager   string
	entryFiles       []string
	configFiles      []string
	buildSystem      string
	testSystem       string
	deploymentTarget string
	notes            []string
}

var knownConfigNames = map[string]struct{}{
	"go.mod":            {},
	"go.sum":            {},
	"package.json":      {},
	"package-lock.json": {},
	"yarn.lock":         {},
	"pnpm-lock.yaml":    {},
	"tsconfig.json":     {},
	"pyproject.toml":    {},
	"requirements.txt":  {},
	"poetry.lock":       {},
	"Pipfile.lock":      {},
	"Cargo.toml":        {},
	"Cargo.lock":        {},
	"pom.xml":           {},
	"build.gradle":      {},
	"build.gradle.kts":  {},
	"composer.json":     {},
	"composer.lock":     {},
	"pubspec.yaml":      {},
	"pubspec.lock":      {},
	"CMakeLists.txt":    {},
	"Makefile":          {},
	"Dockerfile":        {},
	"Taskfile.yml":      {},
	"Taskfile.yaml":     {},
	"angular.json":      {},
	"pytest.ini":        {},
	"conftest.py":       {},
	"Gemfile":           {},
	"Gemfile.lock":      {},
}

func detectConfigFiles(scan scanResult, signals *signalCollector) []string {
	var out []string
	for _, f := range scan.files {
		base := filepath.Base(f)
		if _, ok := knownConfigNames[base]; ok {
			out = append(out, f)
			signals.add("found " + f)
			continue
		}
		if hasAnyPrefix(base, []string{"next.config.", "nuxt.config.", "vite.config.", "webpack.config.", "jest.config.", "vitest.config."}) {
			out = append(out, f)
			signals.add("found " + f)
		}
	}
	sort.Strings(out)
	return dedupeStrings(out)
}

func detectLanguages(scan scanResult, signals *signalCollector) []string {
	scores := map[string]int{}
	for ext, n := range scan.extCounts {
		lang, ok := extToLanguage[ext]
		if !ok {
			continue
		}
		scores[lang] += n * 10
	}

	boost := func(cond bool, language, signal string, score int) {
		if cond {
			scores[language] += score
			signals.add(signal)
		}
	}

	boost(scan.hasBase("go.mod"), "Go", "found go.mod", 40)
	boost(scan.hasBase("package.json"), "JavaScript", "found package.json", 30)
	boost(scan.hasBase("tsconfig.json"), "TypeScript", "found tsconfig.json", 30)
	boost(scan.hasBase("pyproject.toml") || scan.hasBase("requirements.txt"), "Python", "found Python manifest", 40)
	boost(scan.hasBase("Cargo.toml"), "Rust", "found Cargo.toml", 40)
	boost(scan.hasBase("pom.xml") || scan.hasBase("build.gradle") || scan.hasBase("build.gradle.kts"), "Java", "found Java build manifest", 40)
	boost(scan.hasBase("composer.json"), "PHP", "found composer.json", 40)
	boost(scan.hasBase("pubspec.yaml"), "Dart", "found pubspec.yaml", 40)

	type pair struct {
		k string
		v int
	}
	var ranked []pair
	for k, v := range scores {
		if v > 0 {
			ranked = append(ranked, pair{k: k, v: v})
		}
	}
	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].v == ranked[j].v {
			return ranked[i].k < ranked[j].k
		}
		return ranked[i].v > ranked[j].v
	})

	out := make([]string, 0, len(ranked))
	for _, p := range ranked {
		out = append(out, p.k)
	}
	return out
}

func detectFrameworksAndRuntime(scan scanResult, signals *signalCollector) ([]string, string) {
	frameworks := []string{}
	add := func(cond bool, framework, signal string) {
		if cond {
			frameworks = append(frameworks, framework)
			signals.add(signal)
		}
	}

	add(hasBasePrefix(scan, "next.config."), "Next.js", "found next.config.*")
	add(hasBasePrefix(scan, "nuxt.config."), "Nuxt", "found nuxt.config.*")
	add(hasBasePrefix(scan, "vite.config."), "Vite", "found vite.config.*")
	add(hasBasePrefix(scan, "webpack.config."), "Webpack", "found webpack.config.*")
	add(scan.hasBase("angular.json"), "Angular", "found angular.json")
	pubspecPaths := scan.pathsByBase("pubspec.yaml")
	add(len(pubspecPaths) > 0 && strings.Contains(readFileIfExists(scan.root, pubspecPaths[0]), "flutter:"), "Flutter", "found flutter marker in pubspec.yaml")

	deps := readPackageDependencies(scan)
	add(hasDep(deps, "next"), "Next.js", "found next in package.json")
	add(hasDep(deps, "react"), "React", "found react in package.json")
	add(hasDep(deps, "vue"), "Vue", "found vue in package.json")
	add(hasDep(deps, "svelte"), "Svelte", "found svelte in package.json")
	add(hasDep(deps, "express"), "Express", "found express in package.json")
	add(hasDep(deps, "@nestjs/core"), "NestJS", "found @nestjs/core in package.json")

	frameworks = dedupeStrings(frameworks)
	sort.Strings(frameworks)

	runtime := "unknown"
	switch {
	case scan.hasBase("go.mod"):
		runtime = "go"
	case scan.hasBase("package.json"):
		runtime = "node"
	case scan.hasBase("pyproject.toml") || scan.hasBase("requirements.txt") || scan.hasBase("Pipfile.lock"):
		runtime = "python"
	case scan.hasBase("Cargo.toml"):
		runtime = "rust"
	case scan.hasBase("pom.xml") || scan.hasBase("build.gradle") || scan.hasBase("build.gradle.kts"):
		runtime = "jvm"
	case scan.hasBase("composer.json"):
		runtime = "php"
	case scan.hasBase("pubspec.yaml"):
		runtime = "dart"
	case hasExtension(scan, ".cs"):
		runtime = "dotnet"
	}
	if runtime != "unknown" {
		signals.add("runtime inferred from explicit project markers")
	}

	return frameworks, runtime
}

func detectEntryFiles(scan scanResult, languages []string, signals *signalCollector) []string {
	type candidate struct {
		path  string
		score int
	}
	candidates := []candidate{}
	add := func(path string, score int) {
		if !scan.hasFile(path) {
			return
		}
		candidates = append(candidates, candidate{path: path, score: score})
		signals.add("found " + path)
	}

	for _, lang := range languages {
		switch lang {
		case "Go":
			add("main.go", 100)
			for _, f := range scan.files {
				if cmdMainPattern.MatchString(f) {
					add(f, 90)
				} else if strings.HasSuffix(f, "/main.go") {
					add(f, 70)
				}
			}
		case "JavaScript", "TypeScript":
			add("src/main.ts", 95)
			add("src/index.js", 90)
			add("index.js", 85)
			add("app.js", 85)
			add("server.js", 85)
			add("main.ts", 80)
		case "Python":
			add("main.py", 95)
			add("app.py", 90)
			add("manage.py", 90)
			add("wsgi.py", 85)
		case "Rust":
			add("src/main.rs", 100)
		case "Java":
			for _, f := range scan.files {
				if strings.HasPrefix(f, "src/main/java/") && strings.HasSuffix(f, "Application.java") {
					add(f, 95)
				}
			}
		case "C#":
			add("Program.cs", 95)
			add("Startup.cs", 90)
		case "PHP":
			add("public/index.php", 95)
			add("artisan", 90)
		case "Dart":
			add("lib/main.dart", 95)
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].score == candidates[j].score {
			return candidates[i].path < candidates[j].path
		}
		return candidates[i].score > candidates[j].score
	})

	out := make([]string, 0, len(candidates))
	seen := map[string]struct{}{}
	for _, c := range candidates {
		if _, ok := seen[c.path]; ok {
			continue
		}
		seen[c.path] = struct{}{}
		out = append(out, c.path)
	}
	return out
}

func detectPackageManager(scan scanResult, signals *signalCollector) string {
	type rule struct {
		base string
		name string
	}
	rules := []rule{
		{base: "pnpm-lock.yaml", name: "pnpm"},
		{base: "yarn.lock", name: "yarn"},
		{base: "package-lock.json", name: "npm"},
		{base: "go.mod", name: "go modules"},
		{base: "Cargo.lock", name: "cargo"},
		{base: "Cargo.toml", name: "cargo"},
		{base: "poetry.lock", name: "poetry"},
		{base: "Pipfile.lock", name: "pipenv"},
		{base: "requirements.txt", name: "pip"},
		{base: "pyproject.toml", name: "pip"},
		{base: "composer.lock", name: "composer"},
		{base: "composer.json", name: "composer"},
		{base: "Gemfile.lock", name: "bundler"},
		{base: "Gemfile", name: "bundler"},
		{base: "pubspec.lock", name: "pub"},
		{base: "pubspec.yaml", name: "pub"},
	}
	for _, r := range rules {
		if scan.hasBase(r.base) {
			signals.add("found " + r.base)
			return r.name
		}
	}
	return "unknown"
}

func detectBuildSystem(scan scanResult, signals *signalCollector) string {
	type rule struct {
		check  func(scanResult) bool
		name   string
		signal string
	}
	rules := []rule{
		{check: func(s scanResult) bool { return hasBasePrefix(s, "next.config.") }, name: "next", signal: "found next.config.*"},
		{check: func(s scanResult) bool { return hasBasePrefix(s, "nuxt.config.") }, name: "nuxt", signal: "found nuxt.config.*"},
		{check: func(s scanResult) bool { return hasBasePrefix(s, "vite.config.") }, name: "vite", signal: "found vite.config.*"},
		{check: func(s scanResult) bool { return s.hasBase("angular.json") }, name: "angular cli", signal: "found angular.json"},
		{check: func(s scanResult) bool { return hasBasePrefix(s, "webpack.config.") }, name: "webpack", signal: "found webpack.config.*"},
		{check: func(s scanResult) bool { return s.hasBase("CMakeLists.txt") }, name: "cmake", signal: "found CMakeLists.txt"},
		{check: func(s scanResult) bool { return s.hasBase("Makefile") }, name: "make", signal: "found Makefile"},
		{check: func(s scanResult) bool { return s.hasBase("Taskfile.yml") || s.hasBase("Taskfile.yaml") }, name: "task", signal: "found Taskfile"},
		{check: func(s scanResult) bool { return s.hasBase("go.mod") && hasExtension(s, ".go") }, name: "go toolchain", signal: "found go.mod with .go files"},
		{check: func(s scanResult) bool { return s.hasBase("tsconfig.json") }, name: "typescript setup", signal: "found tsconfig.json"},
		{check: func(s scanResult) bool { return s.hasBase("Dockerfile") }, name: "docker", signal: "found Dockerfile"},
	}

	for _, r := range rules {
		if r.check(scan) {
			signals.add(r.signal)
			return r.name
		}
	}
	return "unknown"
}

var (
	cmdMainPattern  = regexp.MustCompile(`^cmd/[^/]+/main\.go$`)
	testFilePattern = regexp.MustCompile(`(?i)\.(test|spec)\.(js|jsx|ts|tsx)$`)
)

func detectTestSystem(scan scanResult, signals *signalCollector) string {
	for _, f := range scan.files {
		if strings.HasSuffix(f, "_test.go") {
			signals.add("found *_test.go")
			return "go test"
		}
	}
	for _, f := range scan.files {
		if testFilePattern.MatchString(f) {
			signals.add("found JS/TS test files")
			if hasBasePrefix(scan, "vitest.config.") {
				signals.add("found vitest.config.*")
				return "vitest"
			}
			if hasBasePrefix(scan, "jest.config.") {
				signals.add("found jest.config.*")
				return "jest"
			}
			return "js/ts tests"
		}
	}
	if scan.hasBase("pytest.ini") || scan.hasBase("conftest.py") {
		signals.add("found pytest markers")
		return "pytest"
	}
	if hasBasePrefix(scan, "vitest.config.") {
		signals.add("found vitest.config.*")
		return "vitest"
	}
	if hasBasePrefix(scan, "jest.config.") {
		signals.add("found jest.config.*")
		return "jest"
	}
	if scan.hasDir("tests") || scan.hasDir("test") {
		signals.add("found tests directory")
		return "present_unknown"
	}
	return "unknown"
}

func detectProjectType(scan scanResult, languages, frameworks []string, signals *signalCollector, notes *[]string) string {
	types := []string{}
	typeSet := map[string]struct{}{}
	addType := func(cond bool, t, signal string) {
		if cond {
			if _, ok := typeSet[t]; !ok {
				types = append(types, t)
				typeSet[t] = struct{}{}
			}
			signals.add(signal)
		}
	}

	hasLang := func(name string) bool {
		for _, l := range languages {
			if l == name {
				return true
			}
		}
		return false
	}
	hasFramework := func(name string) bool {
		for _, f := range frameworks {
			if f == name {
				return true
			}
		}
		return false
	}

	addType(scan.hasBase("go.mod") && hasLang("Go"), "go_project", "go.mod + .go evidence")
	addType(scan.hasBase("package.json") && (hasFramework("React") || hasFramework("Next.js") || hasFramework("Vue") || hasFramework("Svelte")), "web_app", "package.json + web framework marker")
	addType(scan.hasBase("package.json") && (hasLang("JavaScript") || hasLang("TypeScript")), "node_project", "package.json + JS/TS evidence")
	addType((scan.hasBase("pyproject.toml") || scan.hasBase("requirements.txt")) && hasLang("Python"), "python_project", "python manifest + .py evidence")
	addType(scan.hasBase("Cargo.toml") && hasLang("Rust"), "rust_project", "Cargo.toml + .rs evidence")
	addType((scan.hasBase("pom.xml") || scan.hasBase("build.gradle") || scan.hasBase("build.gradle.kts")) && hasLang("Java"), "java_project", "java build manifest + .java evidence")
	addType(scan.hasBase("composer.json") && hasLang("PHP"), "php_project", "composer.json + .php evidence")
	addType(scan.hasBase("pubspec.yaml") && hasLang("Dart") && hasFramework("Flutter"), "flutter_project", "pubspec.yaml + Dart + Flutter evidence")
	addType(scan.hasBase("pubspec.yaml") && hasLang("Dart"), "dart_project", "pubspec.yaml + .dart evidence")
	addType(scan.hasBase("CMakeLists.txt") && (hasLang("C++") || hasLang("C")), "cpp_project", "CMakeLists.txt + C/C++ evidence")

	if len(types) == 0 {
		*notes = append(*notes, "project_type remains unknown because no strong manifest+language combination was found")
		return "unknown"
	}

	contains := func(name string) bool {
		_, ok := typeSet[name]
		return ok
	}

	// Preferred subtype resolution inside a single language stack.
	if contains("flutter_project") {
		return "flutter_project"
	}
	if contains("web_app") {
		return "web_app"
	}
	if contains("dart_project") {
		return "dart_project"
	}
	if contains("node_project") && len(types) == 1 {
		return "node_project"
	}

	// Monorepo only when multiple distinct ecosystem signatures are present.
	if len(types) > 1 {
		*notes = append(*notes, "multiple strong project-type signatures detected")
		signals.add("multiple project-type signatures found")
		return "monorepo"
	}

	sort.Strings(types)
	return types[0]
}

func detectDeploymentTarget(scan scanResult, signals *signalCollector) string {
	if scan.hasBase("Dockerfile") {
		signals.add("found Dockerfile")
		return "container"
	}
	return "unknown"
}

func hasExtension(scan scanResult, ext string) bool {
	_, ok := scan.extCounts[ext]
	return ok
}

func hasBasePrefix(scan scanResult, prefix string) bool {
	for base := range scan.byBase {
		if strings.HasPrefix(base, prefix) {
			return true
		}
	}
	return false
}

func hasAnyPrefix(value string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}

func readPackageDependencies(scan scanResult) map[string]struct{} {
	type pkgJSON struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	out := map[string]struct{}{}
	for _, path := range scan.pathsByBase("package.json") {
		content := readFileIfExists(scan.root, path)
		if content == "" {
			continue
		}
		var p pkgJSON
		if err := json.Unmarshal([]byte(content), &p); err != nil {
			continue
		}
		for k := range p.Dependencies {
			out[k] = struct{}{}
		}
		for k := range p.DevDependencies {
			out[k] = struct{}{}
		}
	}
	return out
}

func hasDep(deps map[string]struct{}, name string) bool {
	_, ok := deps[name]
	return ok
}

func dedupeStrings(in []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, v := range in {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}
