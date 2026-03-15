// Package state handles reading and writing the .feat/ directory state.
package state

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// StateDirName is the name of the state directory.
	StateDirName = ".feat"

	// CurrentFile is the file that stores the current feature path.
	CurrentFile = "CURRENT"

	// ManifestFile stores the path to the manifest being used.
	ManifestFile = "MANIFEST"
)

// State represents the current feature context.
type State struct {
	// FeaturePath is the current feature being worked on.
	FeaturePath string

	// ManifestPath is the path to the manifest file.
	ManifestPath string

	// Timestamp is when the state was last updated.
	Timestamp time.Time
}

// Manager handles state operations for a project.
type Manager struct {
	projectRoot string
	stateDir    string
}

// NewManager creates a state manager for the given project root.
func NewManager(projectRoot string) *Manager {
	return &Manager{
		projectRoot: projectRoot,
		stateDir:    filepath.Join(projectRoot, StateDirName),
	}
}

// Init creates the .feat/ directory if it doesn't exist.
func (m *Manager) Init() error {
	if err := os.MkdirAll(m.stateDir, 0755); err != nil {
		return fmt.Errorf("creating state directory: %w", err)
	}
	return nil
}

// SetCurrent sets the current feature context.
func (m *Manager) SetCurrent(featurePath, manifestPath string) error {
	if err := m.Init(); err != nil {
		return err
	}

	// Write CURRENT file
	currentPath := filepath.Join(m.stateDir, CurrentFile)
	content := fmt.Sprintf("%s\n%s\n%s\n",
		featurePath,
		manifestPath,
		time.Now().Format(time.RFC3339),
	)
	if err := os.WriteFile(currentPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing current file: %w", err)
	}

	return nil
}

// GetCurrent reads the current feature context.
// Returns nil state if no current feature is set.
func (m *Manager) GetCurrent() (*State, error) {
	currentPath := filepath.Join(m.stateDir, CurrentFile)

	data, err := os.ReadFile(currentPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading current file: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) < 1 {
		return nil, fmt.Errorf("invalid state file format")
	}

	state := &State{
		FeaturePath: lines[0],
	}

	if len(lines) >= 2 {
		state.ManifestPath = lines[1]
	}

	if len(lines) >= 3 {
		state.Timestamp, _ = time.Parse(time.RFC3339, lines[2])
	}

	return state, nil
}

// Clear removes the current feature state.
func (m *Manager) Clear() error {
	currentPath := filepath.Join(m.stateDir, CurrentFile)
	if err := os.Remove(currentPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing current file: %w", err)
	}
	return nil
}

// Exists returns true if the .feat/ directory exists.
func (m *Manager) Exists() bool {
	_, err := os.Stat(m.stateDir)
	return !os.IsNotExist(err)
}

// FormatState returns a human-readable representation of the state.
func FormatState(s *State) string {
	if s == nil {
		return "No active feature.\nRun 'feat work <feature-path>' to start working on a feature.\n"
	}

	var output string
	output += fmt.Sprintf("Current feature: %s\n", s.FeaturePath)

	if s.ManifestPath != "" {
		output += fmt.Sprintf("Manifest: %s\n", s.ManifestPath)
	}

	if !s.Timestamp.IsZero() {
		output += fmt.Sprintf("Since: %s\n", s.Timestamp.Format("2006-01-02 15:04:05"))
	}

	return output
}
