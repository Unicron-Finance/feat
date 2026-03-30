package add

import (
	"testing"

	"github.com/lola-the-lobster/feat/internal/manifest"
)

func TestAdd(t *testing.T) {
	// Create a test manifest
	m := &manifest.Manifest{
		Tree: manifest.Tree{
			Name: "test-project",
			Children: map[string]manifest.Node{
				"cli": {
					Files: []string{"cmd/feat/main.go"},
				},
				"manifest": {
					Files: []string{"internal/manifest/manifest.go"},
					Tests: []string{"internal/manifest/manifest_test.go"},
				},
				"auth": {
					Children: map[string]manifest.Node{
						"login": {
							Files: []string{"internal/auth/login.go"},
						},
					},
				},
			},
		},
	}

	t.Run("add file to existing feature", func(t *testing.T) {
		mCopy := copyManifest(m)
		opts := Options{
			FeaturePath: "manifest",
			FilePath:    "internal/manifest/validate.go",
			ManifestDir: "/tmp",
		}

		result, err := Add(mCopy, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.FeaturePath != "manifest" {
			t.Errorf("expected feature path 'manifest', got '%s'", result.FeaturePath)
		}
		if result.FilePath != "internal/manifest/validate.go" {
			t.Errorf("expected file path 'internal/manifest/validate.go', got '%s'", result.FilePath)
		}
		if result.IsTest {
			t.Error("expected IsTest to be false")
		}
		if result.AddedTo != "files" {
			t.Errorf("expected AddedTo 'files', got '%s'", result.AddedTo)
		}

		// Verify manifest was updated
		if len(mCopy.Tree.Children["manifest"].Files) != 2 {
			t.Errorf("expected 2 files in manifest feature, got %d", len(mCopy.Tree.Children["manifest"].Files))
		}
	})

	t.Run("add test file (auto-detected)", func(t *testing.T) {
		mCopy := copyManifest(m)
		opts := Options{
			FeaturePath: "manifest",
			FilePath:    "internal/manifest/validate_test.go",
			ManifestDir: "/tmp",
		}

		result, err := Add(mCopy, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !result.IsTest {
			t.Error("expected IsTest to be true")
		}
		if result.AddedTo != "tests" {
			t.Errorf("expected AddedTo 'tests', got '%s'", result.AddedTo)
		}
	})

	t.Run("error when feature doesn't exist", func(t *testing.T) {
		mCopy := copyManifest(m)
		opts := Options{
			FeaturePath: "nonexistent",
			FilePath:    "some/file.go",
			ManifestDir: "/tmp",
		}

		_, err := Add(mCopy, opts)
		if err == nil {
			t.Error("expected error for nonexistent feature")
		}
	})

	t.Run("error when file already exists in another feature", func(t *testing.T) {
		mCopy := copyManifest(m)
		opts := Options{
			FeaturePath: "cli",
			FilePath:    "cmd/feat/main.go", // Already exists in cli
			ManifestDir: "/tmp",
		}

		_, err := Add(mCopy, opts)
		if err == nil {
			t.Error("expected error when adding duplicate file")
		}
	})

	t.Run("error when adding to boundary", func(t *testing.T) {
		mCopy := copyManifest(m)
		opts := Options{
			FeaturePath: "auth", // auth is a boundary (has children)
			FilePath:    "some/file.go",
			ManifestDir: "/tmp",
		}

		_, err := Add(mCopy, opts)
		if err == nil {
			t.Error("expected error when adding to boundary")
		}
	})

	t.Run("add file to nested feature", func(t *testing.T) {
		mCopy := copyManifest(m)
		opts := Options{
			FeaturePath: "auth/login",
			FilePath:    "internal/auth/login_test.go",
			ManifestDir: "/tmp",
		}

		result, err := Add(mCopy, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.FeaturePath != "auth/login" {
			t.Errorf("expected feature path 'auth/login', got '%s'", result.FeaturePath)
		}
		if !result.IsTest {
			t.Error("expected IsTest to be true")
		}
	})

	t.Run("normalize path", func(t *testing.T) {
		mCopy := copyManifest(m)
		opts := Options{
			FeaturePath: "cli",
			FilePath:    "./cmd/feat/newfile.go",
			ManifestDir: "/tmp",
		}

		result, err := Add(mCopy, opts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.FilePath != "cmd/feat/newfile.go" {
			t.Errorf("expected normalized path 'cmd/feat/newfile.go', got '%s'", result.FilePath)
		}
	})

	t.Run("empty feature path", func(t *testing.T) {
		mCopy := copyManifest(m)
		opts := Options{
			FeaturePath: "",
			FilePath:    "some/file.go",
			ManifestDir: "/tmp",
		}

		_, err := Add(mCopy, opts)
		if err == nil {
			t.Error("expected error for empty feature path")
		}
	})

	t.Run("empty file path", func(t *testing.T) {
		mCopy := copyManifest(m)
		opts := Options{
			FeaturePath: "cli",
			FilePath:    "",
			ManifestDir: "/tmp",
		}

		_, err := Add(mCopy, opts)
		if err == nil {
			t.Error("expected error for empty file path")
		}
	})
}

func TestIsTestFile(t *testing.T) {
	tests := []struct {
		path     string
		pattern  string
		expected bool
	}{
		{"file.go", "_test.go", false},
		{"file_test.go", "_test.go", true},
		{"internal/pkg/file_test.go", "_test.go", true},
		{"_test.go", "_test.go", true},
		{"test.go", "_test.go", false},
		{"test_file.go", "_test.go", false},
		// Custom patterns
		{"file.spec.js", ".spec.js", true},
		{"file.js", ".spec.js", false},
		{"test.py", "test_", false},
	}

	for _, tc := range tests {
		result := isTestFile(tc.path, tc.pattern)
		if result != tc.expected {
			t.Errorf("isTestFile(%q, %q) = %v, expected %v", tc.path, tc.pattern, result, tc.expected)
		}
	}
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"./file.go", "file.go"},
		{"file.go", "file.go"},
		{"./internal/file.go", "internal/file.go"},
		{"internal//file.go", "internal/file.go"},
	}

	for _, tc := range tests {
		result := normalizePath(tc.input)
		if result != tc.expected {
			t.Errorf("normalizePath(%q) = %q, expected %q", tc.input, result, tc.expected)
		}
	}
}

func TestFindFileInManifest(t *testing.T) {
	m := &manifest.Manifest{
		Tree: manifest.Tree{
			Name: "test",
			Children: map[string]manifest.Node{
				"cli": {
					Files: []string{"cmd/main.go"},
				},
				"manifest": {
					Files: []string{"internal/manifest.go"},
					Tests: []string{"internal/manifest_test.go"},
					Children: map[string]manifest.Node{
						"validate": {
							Files: []string{"internal/validate.go"},
						},
					},
				},
			},
		},
	}

	tests := []struct {
		filePath string
		expected string
	}{
		{"cmd/main.go", "cli"},
		{"internal/manifest.go", "manifest"},
		{"internal/manifest_test.go", "manifest"},
		{"internal/validate.go", "manifest/validate"},
		{"nonexistent.go", ""},
	}

	for _, tc := range tests {
		result := findFileInManifest(m, tc.filePath)
		if result != tc.expected {
			t.Errorf("findFileInManifest(%q) = %q, expected %q", tc.filePath, result, tc.expected)
		}
	}
}

// copyManifest creates a deep copy of a manifest for testing.
func copyManifest(m *manifest.Manifest) *manifest.Manifest {
	mCopy := &manifest.Manifest{
		Tree: manifest.Tree{
			Name:     m.Tree.Name,
			Files:    append([]string(nil), m.Tree.Files...),
			Children: make(map[string]manifest.Node),
		},
	}

	var copyNode func(manifest.Node) manifest.Node
	copyNode = func(n manifest.Node) manifest.Node {
		result := manifest.Node{
			Files: append([]string(nil), n.Files...),
			Tests: append([]string(nil), n.Tests...),
		}
		if n.Children != nil {
			result.Children = make(map[string]manifest.Node)
			for name, child := range n.Children {
				result.Children[name] = copyNode(child)
			}
		}
		return result
	}

	for name, child := range m.Tree.Children {
		mCopy.Tree.Children[name] = copyNode(child)
	}

	return mCopy
}

// TestFormatResult verifies the result formatting.
func TestFormatResult(t *testing.T) {
	result := &Result{
		FeaturePath: "manifest",
		FilePath:    "internal/manifest/validate.go",
		IsTest:      false,
		AddedTo:     "files",
	}

	output := FormatResult(result)
	if output == "" {
		t.Error("FormatResult returned empty string")
	}

	// Check that it contains key information
	if !contains(output, "manifest") {
		t.Error("output should contain feature path")
	}
	if !contains(output, "validate.go") {
		t.Error("output should contain file path")
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
