package engine

type ProjectInfo struct {
	ProjectName      string   `json:"project_name"`
	ProjectType      string   `json:"project_type"`
	Languages        []string `json:"languages"`
	Frameworks       []string `json:"frameworks"`
	Runtime          string   `json:"runtime"`
	PackageManager   string   `json:"package_manager"`
	EntryFiles       []string `json:"entry_files"`
	ConfigFiles      []string `json:"config_files"`
	BuildSystem      string   `json:"build_system"`
	TestSystem       string   `json:"test_system"`
	DeploymentTarget string   `json:"deployment_target"`
	Confidence       int      `json:"confidence"`
	Signals          []string `json:"signals"`
	Notes            []string `json:"notes"`
}

type scanResult struct {
	root        string
	projectName string
	files       []string
	fileSet     map[string]struct{}
	dirs        map[string]struct{}
	extCounts   map[string]int
	byBase      map[string][]string
}

type signalCollector struct {
	order []string
	seen  map[string]struct{}
}

func newSignalCollector() *signalCollector {
	return &signalCollector{seen: map[string]struct{}{}}
}

func (s *signalCollector) add(signal string) {
	if signal == "" {
		return
	}
	if _, ok := s.seen[signal]; ok {
		return
	}
	s.seen[signal] = struct{}{}
	s.order = append(s.order, signal)
}

func (s *signalCollector) list() []string {
	out := make([]string, len(s.order))
	copy(out, s.order)
	return out
}
