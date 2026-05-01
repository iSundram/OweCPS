package engine

import "sort"

func DetectProject(rootPath string) (ProjectInfo, error) {
	scan, err := scanRepository(rootPath)
	if err != nil {
		return ProjectInfo{}, err
	}

	signals := newSignalCollector()
	notes := []string{}

	configFiles := detectConfigFiles(scan, signals)
	languages := detectLanguages(scan, signals)
	frameworks, runtime := detectFrameworksAndRuntime(scan, languages, signals)
	entryFiles := detectEntryFiles(scan, languages, signals)
	packageManager := detectPackageManager(scan, languages, signals)
	buildSystem := detectBuildSystem(scan, signals)
	testSystem := detectTestSystem(scan, signals)
	deploymentTarget := detectDeploymentTarget(scan, signals)
	projectType := detectProjectType(scan, languages, frameworks, signals, &notes)

	if len(entryFiles) == 0 {
		notes = append(notes, "no conventional entry file found from deterministic rules")
	}
	if len(languages) > 1 {
		notes = append(notes, "multiple languages detected; ordered by deterministic evidence strength")
	}
	if len(signals.list()) == 0 {
		notes = append(notes, "no useful evidence found")
	}

	confidence := scoreDetection(detected{
		projectType:      projectType,
		languages:        languages,
		frameworks:       frameworks,
		packageManager:   packageManager,
		entryFiles:       entryFiles,
		buildSystem:      buildSystem,
		testSystem:       testSystem,
		deploymentTarget: deploymentTarget,
	}, signals.list())

	sort.Strings(configFiles)
	sort.Strings(notes)

	return ProjectInfo{
		ProjectName:      scan.projectName,
		ProjectType:      projectType,
		Languages:        languages,
		Frameworks:       frameworks,
		Runtime:          runtime,
		PackageManager:   packageManager,
		EntryFiles:       entryFiles,
		ConfigFiles:      configFiles,
		BuildSystem:      buildSystem,
		TestSystem:       testSystem,
		DeploymentTarget: deploymentTarget,
		Confidence:       confidence,
		Signals:          signals.list(),
		Notes:            notes,
	}, nil
}
