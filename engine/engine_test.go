package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFixtureFile(t *testing.T, root, rel, content string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", rel, err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
}

func TestDetectProject_GoService(t *testing.T) {
	root := t.TempDir()
	writeFixtureFile(t, root, "go.mod", "module example.com/app\n")
	writeFixtureFile(t, root, "cmd/api/main.go", "package main\nfunc main(){}\n")
	writeFixtureFile(t, root, "internal/app/app.go", "package app\n")
	writeFixtureFile(t, root, "internal/app/app_test.go", "package app\n")
	writeFixtureFile(t, root, "Dockerfile", "FROM golang:1.22\n")

	info, err := DetectProject(root)
	if err != nil {
		t.Fatalf("DetectProject error: %v", err)
	}

	if info.ProjectType != "go_project" {
		t.Fatalf("project type = %s, want go_project", info.ProjectType)
	}
	if len(info.Languages) == 0 || info.Languages[0] != "Go" {
		t.Fatalf("languages = %v, want Go first", info.Languages)
	}
	if len(info.EntryFiles) == 0 || info.EntryFiles[0] != "cmd/api/main.go" {
		t.Fatalf("entry files = %v, want cmd/api/main.go", info.EntryFiles)
	}
	if info.PackageManager != "go modules" {
		t.Fatalf("package manager = %s, want go modules", info.PackageManager)
	}
	if info.TestSystem != "go test" {
		t.Fatalf("test system = %s, want go test", info.TestSystem)
	}
	if info.DeploymentTarget != "container" {
		t.Fatalf("deployment target = %s, want container", info.DeploymentTarget)
	}
	if info.Confidence < 70 {
		t.Fatalf("confidence = %d, expected >= 70", info.Confidence)
	}
}

func TestDetectProject_NextJS(t *testing.T) {
	root := t.TempDir()
	writeFixtureFile(t, root, "package.json", `{"dependencies":{"next":"14.0.0","react":"18.0.0"}}`)
	writeFixtureFile(t, root, "next.config.js", "module.exports = {}\n")
	writeFixtureFile(t, root, "src/main.ts", "console.log('start')\n")
	writeFixtureFile(t, root, "tsconfig.json", "{}\n")
	writeFixtureFile(t, root, "tests/app.test.ts", "test('x',()=>{})\n")
	writeFixtureFile(t, root, "jest.config.js", "module.exports={}\n")

	info, err := DetectProject(root)
	if err != nil {
		t.Fatalf("DetectProject error: %v", err)
	}

	if info.ProjectType != "web_app" {
		t.Fatalf("project type = %s, want web_app", info.ProjectType)
	}
	if info.Runtime != "node" {
		t.Fatalf("runtime = %s, want node", info.Runtime)
	}
	if info.BuildSystem != "next" {
		t.Fatalf("build system = %s, want next", info.BuildSystem)
	}
	if info.TestSystem != "jest" {
		t.Fatalf("test system = %s, want jest", info.TestSystem)
	}
	if len(info.Frameworks) == 0 {
		t.Fatalf("frameworks should not be empty")
	}
}

func TestDetectProject_Unknown(t *testing.T) {
	root := t.TempDir()
	writeFixtureFile(t, root, "README.txt", "hello\n")

	info, err := DetectProject(root)
	if err != nil {
		t.Fatalf("DetectProject error: %v", err)
	}

	if info.ProjectType != "unknown" {
		t.Fatalf("project type = %s, want unknown", info.ProjectType)
	}
	if info.Confidence != 0 {
		t.Fatalf("confidence = %d, want 0", info.Confidence)
	}
}

func TestDetectProject_FullStack(t *testing.T) {
	root := t.TempDir()
	writeFixtureFile(t, root, "go.mod", "module example.com/fullstack\n")
	writeFixtureFile(t, root, "cmd/api/main.go", "package main\nfunc main(){}\n")
	writeFixtureFile(t, root, "package.json", `{"dependencies":{"react":"18.0.0"}}`)
	writeFixtureFile(t, root, "src/main.ts", "console.log('ui')\n")

	info, err := DetectProject(root)
	if err != nil {
		t.Fatalf("DetectProject error: %v", err)
	}

	if info.ProjectType != "full_stack" {
		t.Fatalf("project type = %s, want full_stack", info.ProjectType)
	}
}

// TestDetectProject_GoLibrary verifies that a Go module without any main.go
// is classified as go_library, not go_project.
func TestDetectProject_GoLibrary(t *testing.T) {
	root := t.TempDir()
	writeFixtureFile(t, root, "go.mod", "module example.com/lib\n")
	writeFixtureFile(t, root, "cobra.go", "package cobra\n")
	writeFixtureFile(t, root, "cobra_test.go", "package cobra\n")

	info, err := DetectProject(root)
	if err != nil {
		t.Fatalf("DetectProject error: %v", err)
	}
	if info.ProjectType != "go_library" {
		t.Fatalf("project type = %s, want go_library", info.ProjectType)
	}
	if info.Runtime != "go" {
		t.Fatalf("runtime = %s, want go", info.Runtime)
	}
	if info.PackageManager != "go modules" {
		t.Fatalf("package manager = %s, want go modules", info.PackageManager)
	}
	if info.TestSystem != "go test" {
		t.Fatalf("test system = %s, want go test", info.TestSystem)
	}
}

// TestDetectProject_RustProject verifies Cargo build system and cargo-test detection.
func TestDetectProject_RustProject(t *testing.T) {
	root := t.TempDir()
	writeFixtureFile(t, root, "Cargo.toml", "[package]\nname = \"myapp\"\n")
	writeFixtureFile(t, root, "Cargo.lock", "")
	writeFixtureFile(t, root, "src/main.rs", "fn main(){}\n")
	writeFixtureFile(t, root, "tests/integration_test.rs", "#[test]\nfn it_works(){}\n")

	info, err := DetectProject(root)
	if err != nil {
		t.Fatalf("DetectProject error: %v", err)
	}
	if info.ProjectType != "rust_project" {
		t.Fatalf("project type = %s, want rust_project", info.ProjectType)
	}
	if info.Runtime != "rust" {
		t.Fatalf("runtime = %s, want rust", info.Runtime)
	}
	if info.BuildSystem != "cargo" {
		t.Fatalf("build system = %s, want cargo", info.BuildSystem)
	}
	if info.TestSystem != "cargo test" {
		t.Fatalf("test system = %s, want cargo test", info.TestSystem)
	}
	if len(info.EntryFiles) == 0 || info.EntryFiles[0] != "src/main.rs" {
		t.Fatalf("entry files = %v, want src/main.rs", info.EntryFiles)
	}
}

// TestDetectProject_JavaMaven verifies maven package manager and build system
// are detected from pom.xml; build.gradle in a sub-directory must not win.
func TestDetectProject_JavaMaven(t *testing.T) {
	root := t.TempDir()
	writeFixtureFile(t, root, "pom.xml", "<project/>")
	writeFixtureFile(t, root, "src/main/java/com/example/App.java", "public class App{}")
	// Simulate Android sub-dir with its own Gemfile (like Flutter Gallery)
	writeFixtureFile(t, root, "android/Gemfile", "source 'https://rubygems.org'")

	info, err := DetectProject(root)
	if err != nil {
		t.Fatalf("DetectProject error: %v", err)
	}
	if info.ProjectType != "java_project" {
		t.Fatalf("project type = %s, want java_project", info.ProjectType)
	}
	if info.Runtime != "jvm" {
		t.Fatalf("runtime = %s, want jvm", info.Runtime)
	}
	if info.PackageManager != "maven" {
		t.Fatalf("package manager = %s, want maven", info.PackageManager)
	}
	if info.BuildSystem != "maven" {
		t.Fatalf("build system = %s, want maven", info.BuildSystem)
	}
}

// TestDetectProject_PythonWithJSTooling ensures a Python project that
// carries a package.json for JS linting does not get classified as node.
func TestDetectProject_PythonWithJSTooling(t *testing.T) {
	root := t.TempDir()
	writeFixtureFile(t, root, "pyproject.toml", "[project]\nname=\"myapp\"\n")
	writeFixtureFile(t, root, "myapp/views.py", "def index(): pass\n")
	writeFixtureFile(t, root, "myapp/models.py", "")
	// biome / eslint tooling
	writeFixtureFile(t, root, "package.json", `{"devDependencies":{"@biomejs/biome":"1.0"}}`)

	info, err := DetectProject(root)
	if err != nil {
		t.Fatalf("DetectProject error: %v", err)
	}
	if info.ProjectType != "python_project" {
		t.Fatalf("project type = %s, want python_project", info.ProjectType)
	}
	if info.Runtime != "python" {
		t.Fatalf("runtime = %s, want python (got %s)", info.Runtime, info.Runtime)
	}
}

// TestDetectProject_FlutterApp verifies runtime=dart and pkg_mgr=pub even
// when an android/Gemfile and android/build.gradle are present.
func TestDetectProject_FlutterApp(t *testing.T) {
	root := t.TempDir()
	writeFixtureFile(t, root, "pubspec.yaml", "name: gallery\ndependencies:\n  flutter:\n    sdk: flutter\n")
	writeFixtureFile(t, root, "pubspec.lock", "")
	writeFixtureFile(t, root, "lib/main.dart", "void main(){}\n")
	writeFixtureFile(t, root, "android/build.gradle", "apply plugin: 'com.android.application'\n")
	writeFixtureFile(t, root, "android/Gemfile", "source 'https://rubygems.org'\n")
	writeFixtureFile(t, root, "test/widget_test.dart", "void main(){}\n")

	info, err := DetectProject(root)
	if err != nil {
		t.Fatalf("DetectProject error: %v", err)
	}
	if info.ProjectType != "flutter_project" {
		t.Fatalf("project type = %s, want flutter_project", info.ProjectType)
	}
	if info.Runtime != "dart" {
		t.Fatalf("runtime = %s, want dart", info.Runtime)
	}
	if info.PackageManager != "pub" {
		t.Fatalf("package manager = %s, want pub", info.PackageManager)
	}
	if len(info.EntryFiles) == 0 || info.EntryFiles[0] != "lib/main.dart" {
		t.Fatalf("entry files = %v, want lib/main.dart", info.EntryFiles)
	}
}

// TestDetectProject_NodeModulesExcluded verifies that node_modules is never
// walked, preventing the signal/config explosion seen on large JS monorepos.
func TestDetectProject_NodeModulesExcluded(t *testing.T) {
	root := t.TempDir()
	writeFixtureFile(t, root, "package.json", `{"dependencies":{"react":"18.0.0"}}`)
	writeFixtureFile(t, root, "next.config.js", "module.exports={}")
	// Simulate an installed dependency tree with its own package.json/tsconfig
	for i := 0; i < 50; i++ {
		writeFixtureFile(t, root, fmt.Sprintf("node_modules/pkg%d/package.json", i), "{}")
		writeFixtureFile(t, root, fmt.Sprintf("node_modules/pkg%d/index.js", i), "")
	}

	info, err := DetectProject(root)
	if err != nil {
		t.Fatalf("DetectProject error: %v", err)
	}
	// signals must not blow up from node_modules content
	if len(info.Signals) > 50 {
		t.Fatalf("signals = %d, expected <= 50; node_modules likely not excluded", len(info.Signals))
	}
	if info.ConfigFiles != nil {
		for _, cf := range info.ConfigFiles {
			if strings.Contains(cf, "node_modules") {
				t.Fatalf("config_files contains node_modules entry: %s", cf)
			}
		}
	}
}
