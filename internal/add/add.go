// Package add handles adding files to existing features in the manifest.
package add

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lola-the-lobster/feat/internal/manifest"
)

// Options configures the add operation.
type Options struct {
	// FeaturePath is the path to the feature (e.g., "manifest" or "auth/login").
	FeaturePath string

	// FilePath is the path to the file to add (relative to manifest directory).
	FilePath string

	// ManifestDir is the directory containing the manifest (for path resolution).
	ManifestDir string

	// ForceTest forces the file to be treated as a test file.
	ForceTest bool
}

// Result contains the outcome of an add operation.
type Result struct {
	// FeaturePath is the feature the file was added to.
	FeaturePath string

	// FilePath is the normalized file path that was added.
	FilePath string

	// IsTest is true if the file was added as a test file.
	IsTest bool

	// AddedTo indicates which list the file was added to ("files" or "tests").
	AddedTo string
}

// Add adds a file to an existing feature.
func Add(m *manifest.Manifest, opts Options) (*Result, error) {
	if opts.FeaturePath == "" {
		return nil, fmt.Errorf("feature path cannot be empty")
	}
	if opts.FilePath == "" {
		return nil, fmt.Errorf("file path cannot be empty")
	}

	// Normalize the file path
	normalizedPath := normalizePath(opts.FilePath)

	// Check if file already exists in any feature
	if existingFeature := findFileInManifest(m, normalizedPath); existingFeature != "" {
		return nil, fmt.Errorf("file already exists in feature '%s': %s", existingFeature, normalizedPath)
	}

	// Navigate to the feature
	parts := splitPath(opts.FeaturePath)
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid feature path: %s", opts.FeaturePath)
	}

	node, parentMap, nodeName, err := navigateToFeature(m, parts)
	if err != nil {
		return nil, err
	}

	// Verify it's a feature (has content), not a boundary
	if !node.IsFeature() && node.Children != nil {
		return nil, fmt.Errorf("'%s' is a boundary, not a feature. Cannot add files to boundaries", opts.FeaturePath)
	}

	// Determine if it's a test file using config pattern
	testPattern := m.Config.GetTestPattern()
	isTest := opts.ForceTest || isTestFile(normalizedPath, testPattern)

	// Add to appropriate list
	if isTest {
		node.Tests = append(node.Tests, normalizedPath)
	} else {
		node.Files = append(node.Files, normalizedPath)
	}

	// Update the node in the parent map
	parentMap[nodeName] = node

	// Check if file exists on disk (warning only)
	fullPath := filepath.Join(opts.ManifestDir, normalizedPath)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		// File doesn't exist yet - this is a warning, not an error
		// The user might be planning to create it
	}

	return &Result{
		FeaturePath: opts.FeaturePath,
		FilePath:    normalizedPath,
		IsTest:      isTest,
		AddedTo:     map[bool]string{true: "tests", false: "files"}[isTest],
	}, nil
}

// navigateToFeature navigates to a feature node and returns it along with its parent map.
func navigateToFeature(m *manifest.Manifest, parts []string) (manifest.Node, map[string]manifest.Node, string, error) {
	if m.Tree.Children == nil {
		return manifest.Node{}, nil, "", fmt.Errorf("manifest has no children")
	}

	current := m.Tree.Children
	var parentMap map[string]manifest.Node
	var nodeName string

	for i, part := range parts {
		node, exists := current[part]
		if !exists {
			return manifest.Node{}, nil, "", fmt.Errorf("feature not found: %s", strings.Join(parts[:i+1], "/"))
		}

		// Last part - this is our target
		if i == len(parts)-1 {
			parentMap = current
			nodeName = part
			return node, parentMap, nodeName, nil
		}

		// Not the last part - must be a boundary with children
		if node.Children == nil {
			return manifest.Node{}, nil, "", fmt.Errorf("'%s' is a feature, cannot navigate deeper", strings.Join(parts[:i+1], "/"))
		}
		current = node.Children
	}

	return manifest.Node{}, nil, "", fmt.Errorf("feature not found")
}

// findFileInManifest searches for a file in all features in the manifest.
func findFileInManifest(m *manifest.Manifest, filePath string) string {
	var finder func(map[string]manifest.Node) string
	finder = func(children map[string]manifest.Node) string {
		for name, node := range children {
			// Check this node's files
			for _, f := range node.Files {
				if f == filePath {
					return name
				}
			}
			// Check this node's tests
			for _, t := range node.Tests {
				if t == filePath {
					return name
				}
			}
			// Recurse into children
			if node.Children != nil {
				if found := finder(node.Children); found != "" {
					return name + "/" + found
				}
			}
		}
		return ""
	}

	return finder(m.Tree.Children)
}

// isTestFile returns true if the file is a test file (matches the pattern).
func isTestFile(path, pattern string) bool {
	return strings.HasSuffix(path, pattern)
}

// normalizePath cleans up the file path.
func normalizePath(path string) string {
	// Remove leading ./ if present
	path = strings.TrimPrefix(path, "./")
	// Clean up any double slashes, etc.
	return filepath.Clean(path)
}

// splitPath splits a path into parts.
func splitPath(path string) []string {
	var parts []string
	for _, p := range strings.Split(path, "/") {
		if p != "" {
			parts = append(parts, p)
		}
	}
	return parts
}

// FormatResult returns a human-readable string representation of the result.
func FormatResult(r *Result) string {
	var output string
	output += fmt.Sprintf("Added file to feature '%s':\n", r.FeaturePath)
	output += fmt.Sprintf("  File: %s\n", r.FilePath)
	output += fmt.Sprintf("  Type: %s\n", r.AddedTo)
	return output
}
