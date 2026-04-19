package engine

import (
"os"
"path/filepath"
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

if info.ProjectType != "monorepo" {
t.Fatalf("project type = %s, want monorepo due to overlapping strong signatures", info.ProjectType)
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
