package engine

func scoreDetection(d detected, signals []string) int {
	if len(signals) == 0 {
		return 0
	}
	score := 0
	if d.projectType != "unknown" {
		score += 30
	}
	if len(d.languages) > 0 {
		score += 20
	}
	if len(d.entryFiles) > 0 {
		score += 15
	}
	if len(d.frameworks) > 0 {
		score += 10
	}
	if d.packageManager != "unknown" {
		score += 10
	}
	if d.buildSystem != "unknown" {
		score += 5
	}
	if d.testSystem != "unknown" {
		score += 5
	}
	if d.deploymentTarget != "unknown" {
		score += 5
	}
	if score > 100 {
		return 100
	}
	return score
}
