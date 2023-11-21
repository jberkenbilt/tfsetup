package tfsetup

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const tpl = `# Hello, {{.Config.name}} at {{.Path}}.
{{if .Project.things
}}# Things:
{{- range .Project.things }}
# • {{.}}
{{- end}}
{{end -}}
`

func TestGenerate(t *testing.T) {
	configContextBytes := []byte(`{"name": "Potato"}`)
	projectContextBytes := []byte(`{"things": ["a", "b"]}`)
	outBytes, err := generate(projectContextBytes, configContextBytes, tpl, "x/y/z")
	if err != nil {
		t.Errorf(err.Error())
	}
	out := string(outBytes)
	exp := `# Hello, Potato at x/y/z.
# Things:
# • a
# • b
`
	if out != exp {
		t.Errorf("wrong result: %v", out)
	}
}

func TestRun(t *testing.T) {
	_, err := Run(false)
	if !strings.Contains(err.Error(), "tfsetup-context.json") {
		t.Errorf("didn't fail when not found: %v", err)
	}
	dir := t.TempDir()
	projectDir := filepath.Join(dir, "ab/cd/ef")
	configDir := filepath.Join(dir, filepath.Join("ab", configPath))
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer func() {
		_ = os.Chdir(cwd)
	}()
	configContextBytes := []byte(`{"name": "Potato"}`)
	projectContextBytes := []byte(`{"things": ["a", "b"]}`)
	err = errors.Join(
		os.MkdirAll(projectDir, 0777),
		os.MkdirAll(configDir, 0777),
		os.Chdir(projectDir),
		os.WriteFile(filepath.Join(configDir, configTemplate), []byte(tpl), 0777),
		os.WriteFile(filepath.Join(configDir, configContextFile), configContextBytes, 0777),
		os.WriteFile(projectContextFile, projectContextBytes, 0777),
	)
	if err != nil {
		t.Fatalf(err.Error())
	}
	ok, err := Run(false)
	if ok || err != nil {
		t.Errorf("wrong result: %v %v", ok, err)
	}
	ok, err = Run(true)
	if ok || err != nil {
		t.Errorf("wrong result: %v %v", ok, err)
	}
	resultBytes, err := os.ReadFile(SetupFile)
	if err != nil {
		t.Errorf(err.Error())
	}
	result := string(resultBytes)
	exp := `# Hello, Potato at cd/ef.
# Things:
# • a
# • b
`
	if result != exp {
		t.Errorf("wrong results: %v", result)
	}
	ok, err = Run(false)
	if !ok || err != nil {
		t.Errorf("wrong result: %v %v", ok, err)
	}
	projectContextBytes = []byte(`{"things": ["a", "b", "c"]}`)
	err = os.WriteFile(projectContextFile, projectContextBytes, 0777)
	if err != nil {
		t.Fatalf(err.Error())
	}
	ok, err = Run(false)
	if ok || err != nil {
		t.Errorf("wrong result: %v %v", ok, err)
	}
	ok, err = Run(true)
	if ok || err != nil {
		t.Errorf("wrong result: %v %v", ok, err)
	}
	resultBytes, err = os.ReadFile(SetupFile)
	if err != nil {
		t.Errorf(err.Error())
	}
	result = string(resultBytes)
	exp = `# Hello, Potato at cd/ef.
# Things:
# • a
# • b
# • c
`
	if result != exp {
		t.Errorf("wrong results: %v", result)
	}
	ok, err = Run(true)
	if !ok || err != nil {
		t.Errorf("wrong result: %v %v", ok, err)
	}
}
