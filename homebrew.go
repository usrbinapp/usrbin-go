package usrbin

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/pkg/errors"
)

type HomebrewExternalPackageManager struct {
	formula string
}

type homebrewInfoOutput struct {
	Installed []struct {
		Version string `json:"version"`
	} `json:"installed"`
}

func NewHomebrewExternalPackageManager(formula string) ExternalPackageManager {
	return HomebrewExternalPackageManager{
		formula: formula,
	}
}

func (m HomebrewExternalPackageManager) UpgradeCommand() string {
	return fmt.Sprintf("brew upgrade %s", m.formula)
}

// IsInstalled will return true if the formula is installed using homebrew
func (m HomebrewExternalPackageManager) IsInstalled() (bool, error) {
	path, err := exec.LookPath("brew")
	if err != nil {
		// we just assume that it wasn't installed via brew if there's no brew command
		return false, nil
	}

	out, err := exec.Command(
		path,
		"info",
		m.formula,
		"--json",
	).Output()
	if err != nil {
		return false, errors.Wrap(err, "exec brew")
	}

	unmarshaled := []homebrewInfoOutput{}
	if err := json.Unmarshal(out, &unmarshaled); err != nil {
		return false, errors.Wrap(err, "unmarshal brew output")
	}

	if len(unmarshaled) == 0 {
		return false, nil
	}

	if len(unmarshaled[0].Installed) == 0 {
		return false, nil
	}

	return true, nil
}
