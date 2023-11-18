package tfsetup

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/jberkenbilt/tfsetup/util"
	"os"
	"path/filepath"
	"slices"
	"text/template"
)

const projectContextFile = "tfsetup-context.json"
const configPath = "tfsetup-config"
const configContextFile = "context.json"
const configTemplate = "setup.tmpl"
const SetupFile = "setup.tf"

type templateContext struct {
	Config  any
	Project any
	Path    string
}

func generate(
	projectContextBytes []byte,
	configContextBytes []byte,
	configTemplate string,
	relPath string,
) ([]byte, error) {
	var projectContext any
	err := json.Unmarshal(projectContextBytes, &projectContext)
	if err != nil {
		return nil, fmt.Errorf("decode project context: %w", err)
	}
	var configContext any
	err = json.Unmarshal(configContextBytes, &configContext)
	if err != nil {
		return nil, fmt.Errorf("decode config context: %w", err)
	}
	tpl, err := template.New("setup").Parse(configTemplate)
	fullContext := templateContext{
		Config:  configContext,
		Project: projectContext,
		Path:    relPath,
	}
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}
	var buf bytes.Buffer
	err = tpl.Execute(&buf, fullContext)
	if err != nil {
		return nil, fmt.Errorf("evaluate template: %w", err)
	}
	return buf.Bytes(), nil
}

// Run checks or updates the setup file. The boolean return value indicates
// whether the file was already up-to-date.
func Run(allowChanges bool) (bool, error) {
	projectContextBytes, err := os.ReadFile(projectContextFile)
	if err != nil {
		return false, fmt.Errorf("read %s: %w", projectContextFile, err)
	}
	configPath, relPath, err := util.FindDir(configPath)
	if err != nil {
		return false, fmt.Errorf("unable to find %s above current directory: %w", configPath, err)
	}
	configContextBytes, err := os.ReadFile(filepath.Join(configPath, configContextFile))
	if err != nil {
		return false, fmt.Errorf("reading %s from %s: %w", configContextFile, configPath, err)
	}
	configTemplateBytes, err := os.ReadFile(filepath.Join(configPath, configTemplate))
	if err != nil {
		return false, fmt.Errorf("reading %s from %s: %w", configTemplate, configPath, err)
	}
	// It's okay if the setup file is missing at this point.
	origSetupBytes, _ := os.ReadFile(SetupFile)
	setupBytes, err := generate(projectContextBytes, configContextBytes, string(configTemplateBytes), relPath)
	if err != nil {
		return false, err
	}
	if slices.Equal(origSetupBytes, setupBytes) {
		return true, nil
	}
	if !allowChanges {
		return false, nil
	}
	_ = os.Remove(SetupFile)
	err = os.WriteFile(SetupFile, setupBytes, 0444)
	if err != nil {
		return false, fmt.Errorf("error writing %s: %w", SetupFile, err)
	}
	return false, nil
}
