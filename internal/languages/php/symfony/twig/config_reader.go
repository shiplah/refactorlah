package twig

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type ConfigReader struct{}

func (r ConfigReader) Read(projectRoot string) (PathConfiguration, error) {
	roots := map[string]PathRoot{}

	yamlRoots, err := r.readYamlRoots(filepath.Join(projectRoot, "config", "packages", "twig.yaml"))
	if err != nil {
		return PathConfiguration{}, err
	}
	for _, root := range yamlRoots {
		roots[pathRootKey(root)] = root
	}

	phpRoots, err := r.readPhpRoots(filepath.Join(projectRoot, "config", "packages", "twig.php"))
	if err != nil {
		return PathConfiguration{}, err
	}
	for _, root := range phpRoots {
		roots[pathRootKey(root)] = root
	}

	if len(roots) == 0 {
		if info, err := os.Stat(filepath.Join(projectRoot, "templates")); err == nil && info.IsDir() {
			root := PathRoot{Path: "templates"}
			roots[pathRootKey(root)] = root
		}
	}

	result := PathConfiguration{Roots: make([]PathRoot, 0, len(roots))}
	for _, root := range roots {
		result.Roots = append(result.Roots, root)
	}

	return result, nil
}

func (r ConfigReader) readYamlRoots(path string) ([]PathRoot, error) {
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()

	defaultPathPattern := regexp.MustCompile(`^\s{2}default_path:\s*['"]?%kernel\.project_dir%/([^'"]+)['"]?\s*$`)
	pathPattern := regexp.MustCompile(`^\s{4}['"]?%kernel\.project_dir%/([^'"]+)['"]?\s*:\s*([A-Za-z0-9_]+)\s*$`)

	var roots []PathRoot
	inTwigBlock := false
	inPathsBlock := false
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "twig:" {
			inTwigBlock = true
			inPathsBlock = false
			continue
		}
		if inTwigBlock && len(line) > 0 && line[0] != ' ' && line[0] != '\t' {
			inTwigBlock = false
			inPathsBlock = false
		}
		if !inTwigBlock {
			continue
		}

		if matches := defaultPathPattern.FindStringSubmatch(line); matches != nil {
			roots = append(roots, PathRoot{Path: cleanTwigRootPath(matches[1])})
			continue
		}
		if regexp.MustCompile(`^\s{2}paths:\s*$`).MatchString(line) {
			inPathsBlock = true
			continue
		}
		if inPathsBlock {
			if matches := pathPattern.FindStringSubmatch(line); matches != nil {
				roots = append(roots, PathRoot{Path: cleanTwigRootPath(matches[1]), Namespace: matches[2]})
				continue
			}
			if regexp.MustCompile(`^\s{2}\S`).MatchString(line) {
				inPathsBlock = false
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return roots, nil
}

func (r ConfigReader) readPhpRoots(path string) ([]PathRoot, error) {
	source, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	defaultPathPattern := regexp.MustCompile(`->defaultPath\(\s*['"]%kernel\.project_dir%/([^'"]+)['"]\s*\)`)
	pathPattern := regexp.MustCompile(`->path\(\s*['"]%kernel\.project_dir%/([^'"]+)['"]\s*,\s*['"]([^'"]+)['"]\s*\)`)

	var roots []PathRoot
	for _, matches := range defaultPathPattern.FindAllStringSubmatch(string(source), -1) {
		roots = append(roots, PathRoot{Path: cleanTwigRootPath(matches[1])})
	}
	for _, matches := range pathPattern.FindAllStringSubmatch(string(source), -1) {
		roots = append(roots, PathRoot{Path: cleanTwigRootPath(matches[1]), Namespace: matches[2]})
	}

	return roots, nil
}

func cleanTwigRootPath(path string) string {
	return strings.Trim(strings.ReplaceAll(path, "\\", "/"), "/")
}

func pathRootKey(root PathRoot) string {
	return root.Path + "|" + root.Namespace
}
