OweCode Project Identifier Engine Prompt

Build a rule-based project identification engine for OweCode.

This engine must inspect a repository and determine what kind of project it is, which language it uses, what the main entry file is, and what supporting tools or frameworks are present.

Core Requirement

Do not use AI, embeddings, probabilistic prediction, or natural-language guessing.

Use only:

file names

directory names

file extensions

manifest files

lock files

framework config files

explicit markers inside known files

deterministic rules


Main Job

Create an engine that can take a project folder and return structured identification data such as:

project type

primary language

secondary languages

framework or runtime

main entry file

package manager

build system

test system

deployment style

confidence score

evidence list


Engine Design Goals

The engine should be:

deterministic

fast

explainable

easy to extend

easy to debug

safe when information is incomplete


Required Flow

The engine should work in this order:

1. Scan the repository tree.


2. Collect all file names, extensions, and folder names.


3. Detect manifest and config files.


4. Detect language from extensions and manifests.


5. Detect framework or runtime from file patterns.


6. Detect main entry file using language-specific rules.


7. Detect package manager and build system.


8. Detect test setup.


9. Produce a structured JSON result.


10. Attach evidence strings for every conclusion.



Output Contract

Return a structured object like this:

{
  "project_name": "unknown",
  "project_type": "unknown",
  "languages": [],
  "frameworks": [],
  "runtime": "unknown",
  "package_manager": "unknown",
  "entry_files": [],
  "config_files": [],
  "build_system": "unknown",
  "test_system": "unknown",
  "deployment_target": "unknown",
  "confidence": 0,
  "signals": [],
  "notes": []
}

Rules for Identification

1. Project Type

Use strong evidence only. Examples:

go.mod + .go files → Go project

package.json + React/Next/Vue/Svelte markers → JavaScript or TypeScript web app

pyproject.toml / requirements.txt + .py files → Python project

Cargo.toml + .rs files → Rust project

pom.xml / build.gradle + .java files → Java project

composer.json + .php files → PHP project

pubspec.yaml + .dart files → Dart or Flutter project

CMakeLists.txt + .cpp files → C/C++ project


2. Language Detection

Detect language by file extension and supporting manifest evidence. Examples:

.go → Go

.js, .mjs, .cjs → JavaScript

.ts, .tsx → TypeScript

.py → Python

.rs → Rust

.java → Java

.kt, .kts → Kotlin

.dart → Dart

.php → PHP

.rb → Ruby

.cs → C#

.cpp, .cc, .cxx, .hpp, .h → C or C++

.swift → Swift

.scala → Scala

.lua → Lua

.sh, .bash → Shell

.html, .css, .scss → frontend web assets


If multiple languages exist, rank them by strength of evidence.

3. Main Entry File

Use conventional entry-point rules. Examples:

Go: main.go, especially root main.go or cmd/*/main.go

Node.js: index.js, app.js, server.js, src/index.js, src/main.ts

Python: main.py, app.py, manage.py, wsgi.py

Rust: src/main.rs

Java: src/main/java/.../Application.java

C#: Program.cs, Startup.cs

PHP: public/index.php, artisan

Dart/Flutter: lib/main.dart


If more than one entry file exists, return all likely candidates in confidence order.

4. Package Manager

Detect package manager from explicit files. Examples:

package-lock.json → npm

yarn.lock → yarn

pnpm-lock.yaml → pnpm

go.mod → Go modules

Cargo.lock → Cargo

poetry.lock → Poetry

Pipfile.lock → Pipenv

composer.lock → Composer

Gemfile.lock → Bundler

pubspec.lock → pub


5. Build System

Detect build system from known markers. Examples:

Makefile → make

Dockerfile → docker

Taskfile.yml → task

CMakeLists.txt → cmake

vite.config.* → vite

next.config.* → next

nuxt.config.* → nuxt

angular.json → angular cli

webpack.config.* → webpack

tsconfig.json → typescript setup

go.mod + Go files → Go toolchain


6. Test System

Detect test setup from file and directory patterns. Examples:

*_test.go → Go tests

*.test.js, *.spec.js, *.test.ts, *.spec.ts → JS or TS tests

tests/ or test/ directories → test suite present

pytest.ini, conftest.py → pytest

jest.config.* → jest

vitest.config.* → vitest


If only a test folder exists without actual test files, return present_unknown.

Evidence Rules

Every detection must include a signal. Examples:

found go.mod

found main.go

found cmd/app/main.go

found package.json

found next.config.js

found src/app.tsx

found pytest.ini

found *_test.go


Confidence Scoring

Use a numeric score from 0 to 100. Suggested scale:

90–100: manifest + entry file + framework marker

70–89: strong manifest or multiple strong indicators

40–69: partial evidence

1–39: weak evidence only

0: no useful evidence


Handling Ambiguity

When the engine cannot be certain:

return unknown

do not guess

do not infer beyond evidence

add a note explaining the ambiguity


Repository Patterns to Support

The engine should handle:

single-project repositories

monorepos

backend-only projects

frontend-only projects

full-stack projects

libraries

CLI tools

services

containers

generated codebases


Suggested Internal Modules

Split the engine into small deterministic modules:

scanner

manifest detector

language detector

framework detector

entry-point detector

package-manager detector

build-system detector

test detector

scorer

result formatter


Implementation Expectations

The final engine should expose one clean function, for example:

DetectProject(rootPath string) (ProjectInfo, error)

Where ProjectInfo contains all detected fields and evidence.

Final Output Rules

JSON only

no markdown in the runtime output

no AI text generation in the detector

no assumptions without evidence

stable and repeatable results


Development Priority

Build in this order:

1. scanner


2. file pattern rules


3. language detection


4. manifest detection


5. entry file detection


6. framework detection


7. package manager detection


8. test detection


9. confidence scoring


10. output formatting



Success Criteria

The engine is successful when it can reliably identify common projects like:

Go CLI tools

Go web servers

Node.js apps

TypeScript apps

Python apps

Rust crates

Java apps

PHP apps

Flutter apps

Dockerized services


and return accurate JSON without AI.
