package upgrade

import (
	"fmt"
	"strconv"
	"strings"
)

// VersionComponents version number components
type VersionComponents struct {
	Major int
	Minor int
	Patch int
	Pre   string
}

// parseVersion parse version string into components
func parseVersion(version string) (*VersionComponents, error) {
	// Clean version string, remove possible invalid characters
	version = strings.TrimSpace(version)
	if idx := strings.Index(version, " "); idx != -1 {
		// If contains spaces, take the first part
		version = version[:idx]
	}
	// Remove version prefix
	// First check if starts with "version" or "Version"
	if strings.HasPrefix(strings.ToLower(version), "version") {
		version = version[7:] // Skip the 7 characters of "version"
	}
	// Then remove "v" or "V" prefix
	version = strings.TrimPrefix(version, "v")
	version = strings.TrimPrefix(version, "V")

	// Handle pre-release version
	pre := ""
	if idx := strings.IndexAny(version, "-_"); idx != -1 {
		pre = version[idx+1:]
		version = version[:idx]
	}

	// Split version number components
	parts := strings.Split(version, ".")
	if len(parts) < 1 || len(parts) > 3 {
		return nil, fmt.Errorf("invalid version format: %s", version)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid major version: %s", parts[0])
	}

	minor := 0
	if len(parts) > 1 {
		minor, err = strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid minor version: %s", parts[1])
		}
	}

	patch := 0
	if len(parts) > 2 {
		// Handle possible build information
		patchPart := parts[2]
		if idx := strings.IndexAny(patchPart, "+_"); idx != -1 {
			patchPart = patchPart[:idx]
		}
		patch, err = strconv.Atoi(patchPart)
		if err != nil {
			return nil, fmt.Errorf("invalid patch version: %s", parts[2])
		}
	}

	return &VersionComponents{
		Major: major,
		Minor: minor,
		Patch: patch,
		Pre:   pre,
	}, nil
}

// compareVersions compare two version numbers, returns 1 if v1 > v2, 0 if equal, -1 if v1 < v2
func compareVersions(v1, v2 string) (int, error) {
	vc1, err := parseVersion(v1)
	if err != nil {
		return 0, err
	}

	vc2, err := parseVersion(v2)
	if err != nil {
		return 0, err
	}

	// Compare major version
	if vc1.Major > vc2.Major {
		return 1, nil
	} else if vc1.Major < vc2.Major {
		return -1, nil
	}

	// Compare minor version
	if vc1.Minor > vc2.Minor {
		return 1, nil
	} else if vc1.Minor < vc2.Minor {
		return -1, nil
	}

	// Compare patch version
	if vc1.Patch > vc2.Patch {
		return 1, nil
	} else if vc1.Patch < vc2.Patch {
		return -1, nil
	}

	// Compare pre-release version
	if vc1.Pre == "" && vc2.Pre != "" {
		return 1, nil // Release version is higher than pre-release version
	} else if vc1.Pre != "" && vc2.Pre == "" {
		return -1, nil
	} else if vc1.Pre > vc2.Pre {
		return 1, nil
	} else if vc1.Pre < vc2.Pre {
		return -1, nil
	}

	return 0, nil
}

// needsUpgrade check if upgrade is needed
func needsUpgrade(currentVersion, latestVersion string) (bool, error) {
	cmp, err := compareVersions(latestVersion, currentVersion)
	if err != nil {
		return false, err
	}
	return cmp > 0, nil
}
