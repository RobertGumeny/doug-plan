package layout

import "path/filepath"

const (
	PlanDirName    = "plan"
	ConfigFileName = "doug-plan.yaml"
	ActiveStepFile = "ACTIVE_STEP.md"
)

func DougDir(projectRoot string) string {
	return filepath.Join(projectRoot, ".doug")
}

func PlanDir(projectRoot string) string {
	return filepath.Join(DougDir(projectRoot), PlanDirName)
}

func ConfigPath(projectRoot string) string {
	return filepath.Join(PlanDir(projectRoot), ConfigFileName)
}

func ActiveStepPath(projectRoot string) string {
	return filepath.Join(PlanDir(projectRoot), ActiveStepFile)
}

func LogsDir(projectRoot string) string {
	return filepath.Join(PlanDir(projectRoot), "logs")
}

func EpicsDir(projectRoot string) string {
	return filepath.Join(PlanDir(projectRoot), "epics")
}

func ManifestPath(projectRoot string) string {
	return filepath.Join(PlanDir(projectRoot), "manifest.yaml")
}
