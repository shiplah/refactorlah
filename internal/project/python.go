package project

var pythonProjectMarkers = []string{
	"pyproject.toml",
	"setup.cfg",
	"tox.ini",
	"pytest.ini",
	"mypy.ini",
	".mypy.ini",
	"ruff.toml",
	".ruff.toml",
}

func FindPythonRootForPaths(projectRoot string, paths []string) (string, bool, error) {
	return FindMarkerRootForPaths(projectRoot, paths, pythonProjectMarkers)
}
