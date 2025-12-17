package upgrade

import (
	"strings"
	"testing"
)

// TestParseUpgradeList test parseUpgradeList function
func TestParseUpgradeList(t *testing.T) {
	// Use the provided list.txt content from user
	listContent := `a4c9f9295600b17ba118378589aac75b4df8968285ddc2d174a7e3056708f93b  windows-amd64-dlpcli.exe.v1.0.1 
 a4c9f9295600b17ba118378589aac75b4df8968285ddc2d174a7e3056708f93b  windows-amd64-dlpcli.exe.v1.0.1 
 f733d6633183770b2f9f8aa9dd188d4c5ef34d0c5767658acb342b1c7e2776e9  windows-amd64-dlpcli.exe.v1.0.2 
 977fc9f50915d9db873cbe7f6496e54fd90fef30911fc4beb01d8634ccff1f19  windows-amd64-dlpcli.exe.v1.0.3 
 a4c9f9295600b17ba118378589aac75b4df8968285ddc2d174a7e3056708f93b  windows-amd64-dlpcli.exe.v1.0.1 
 f733d6633183770b2f9f8aa9dd188d4c5ef34d0c5767658acb342b1c7e2776e9  windows-amd64-dlpcli.exe.v1.0.2 
 977fc9f50915d9db873cbe7f6496e54fd90fef30911fc4beb01d8634ccff1f19  windows-amd64-dlpcli.exe.v1.0.300`

	upgradeMap, err := parseUpgradeList([]byte(listContent))
	if err != nil {
		t.Fatalf("parseUpgradeList failed: %v", err)
	}

	// Check if 4 unique upgrade package entries are correctly parsed
	if len(upgradeMap) != 4 {
		t.Fatalf("expected 4 unique upgrade packages, got %d", len(upgradeMap))
	}

	// Check if each upgrade package is correctly parsed
	expectedPackages := []string{
		"windows-amd64-dlpcli.exe.v1.0.1",
		"windows-amd64-dlpcli.exe.v1.0.2",
		"windows-amd64-dlpcli.exe.v1.0.3",
		"windows-amd64-dlpcli.exe.v1.0.300",
	}

	for _, pkg := range expectedPackages {
		if _, exists := upgradeMap[pkg]; !exists {
			t.Errorf("expected upgrade package %s not found", pkg)
		}
	}
}

// TestVersionExtraction test version extraction from filename
func TestVersionExtraction(t *testing.T) {
	filename := "windows-amd64-dlpcli.exe.v1.0.300"
	versionPrefix := ".v"
	versionIndex := strings.LastIndex(filename, versionPrefix)
	if versionIndex == -1 {
		t.Fatalf("version prefix %s not found in filename %s", versionPrefix, filename)
	}

	version := filename[versionIndex+len(versionPrefix):]
	expectedVersion := "1.0.300"
	if version != expectedVersion {
		t.Errorf("expected version %s, got %s", expectedVersion, version)
	}
}

// TestCompareVersions test version comparison functionality
func TestCompareVersions(t *testing.T) {
	tests := []struct {
		v1       string
		v2       string
		expected int
	}{
		{"1.0.1", "1.0.2", -1},
		{"1.0.3", "1.0.2", 1},
		{"1.0.0", "1.0.0", 0},
		{"2.0.0", "1.9.9", 1},
		{"1.0.300", "1.0.3", 1},
		{"1.0.3", "1.0.300", -1},
	}

	for _, test := range tests {
		result, err := compareVersions(test.v1, test.v2)
		if err != nil {
			t.Fatalf("compareVersions(%q, %q) failed: %v", test.v1, test.v2, err)
		}
		if result != test.expected {
			t.Errorf("compareVersions(%q, %q): expected %d, got %d", test.v1, test.v2, test.expected, result)
		}
	}
}

// TestFindLatestVersion test finding latest version from upgrade package list
func TestFindLatestVersion(t *testing.T) {
	// 使用用户提供的list.txt内容
	listContent := `a4c9f9295600b17ba118378589aac75b4df8968285ddc2d174a7e3056708f93b  windows-amd64-dlpcli.exe.v1.0.1 
 a4c9f9295600b17ba118378589aac75b4df8968285ddc2d174a7e3056708f93b  windows-amd64-dlpcli.exe.v1.0.1 
 f733d6633183770b2f9f8aa9dd188d4c5ef34d0c5767658acb342b1c7e2776e9  windows-amd64-dlpcli.exe.v1.0.2 
 977fc9f50915d9db873cbe7f6496e54fd90fef30911fc4beb01d8634ccff1f19  windows-amd64-dlpcli.exe.v1.0.3 
 a4c9f9295600b17ba118378589aac75b4df8968285ddc2d174a7e3056708f93b  windows-amd64-dlpcli.exe.v1.0.1 
 f733d6633183770b2f9f8aa9dd188d4c5ef34d0c5767658acb342b1c7e2776e9  windows-amd64-dlpcli.exe.v1.0.2 
 977fc9f50915d9db873cbe7f6496e54fd90fef30911fc4beb01d8634ccff1f19  windows-amd64-dlpcli.exe.v1.0.300`

	upgradeMap, err := parseUpgradeList([]byte(listContent))
	if err != nil {
		t.Fatalf("parseUpgradeList failed: %v", err)
	}

	// Extract latest version from upgrade list (simulate logic in CheckUpgrade)
	var latestVersion string
	for filename := range upgradeMap {
		versionPrefix := ".v"
		versionIndex := strings.LastIndex(filename, versionPrefix)
		if versionIndex == -1 {
			continue
		}

		version := filename[versionIndex+len(versionPrefix):]
		if version == "" {
			continue
		}

		if latestVersion == "" {
			latestVersion = version
		} else {
			cmp, err := compareVersions(version, latestVersion)
			if err == nil && cmp > 0 {
				latestVersion = version
			}
		}
	}

	// Verify if latest version is 1.0.300
	expectedLatestVersion := "1.0.300"
	if latestVersion != expectedLatestVersion {
		t.Errorf("expected latest version %s, got %s", expectedLatestVersion, latestVersion)
	}
}
