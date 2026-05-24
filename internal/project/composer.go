package project

func FindComposerRootForPaths(projectRoot string, paths []string) (string, bool, error) {
	return FindMarkerRootForPaths(projectRoot, paths, []string{"composer.json"})
}
