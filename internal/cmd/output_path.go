package cmd

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	versionDirPattern  = regexp.MustCompile(`^\d+$`)
	invalidPathPattern = regexp.MustCompile(`[\\/:*?"<>|]`)
	spacePattern       = regexp.MustCompile(`\s+`)
)

func BuildDefaultOutputDir(cwd, appID, appName, version string) string {
	appSegment := appID
	if sanitizedName := sanitizePathSegment(appName); sanitizedName != "" {
		appSegment = appID + "-" + sanitizedName
	}

	version = sanitizePathSegment(version)
	if version == "" {
		version = "unknown"
	}

	return filepath.Join(cwd, "output", appSegment, version)
}

func DetectVersionFromInput(input string) string {
	firstInput := firstInputPath(input)
	if firstInput == "" {
		return "unknown"
	}

	targetPath := firstInput
	if info, err := os.Stat(firstInput); err == nil {
		if !info.IsDir() {
			targetPath = filepath.Dir(firstInput)
		}
	} else {
		targetPath = filepath.Dir(firstInput)
	}

	version := filepath.Base(filepath.Clean(targetPath))
	if versionDirPattern.MatchString(version) {
		return version
	}

	return "unknown"
}

func sanitizePathSegment(value string) string {
	value = strings.TrimSpace(value)
	value = invalidPathPattern.ReplaceAllString(value, "_")
	value = spacePattern.ReplaceAllString(value, " ")
	value = strings.Trim(value, ". ")
	return value
}

func firstInputPath(input string) string {
	parts := strings.Split(input, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			return part
		}
	}
	return ""
}
