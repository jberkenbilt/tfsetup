package tfsetup

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const input = `# Hello, {{.Config.name}} at {{.Path}}.
{{if .Project.things
}}# Things:
{{- range .Project.things }}
# • {{.}}, {{.|upper}}
{{- end}}
{{end -}}
`

func TestRender(t *testing.T) {
	configContextBytes := []byte(`{"name": "Potato"}`)
	projectContextBytes := []byte(`{"things": ["a", "b"]}`)
	tpl, err := newTemplateContext(projectContextBytes, configContextBytes, "x/y/z")
	if err != nil {
		t.Fatal(err.Error())
	}
	outBytes, err := tpl.render("input", []byte(input))
	if err != nil {
		t.Error(err.Error())
	}
	out := string(outBytes)
	exp := `# Hello, Potato at x/y/z.
# Things:
# • a, A
# • b, B
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
		t.Fatal(err.Error())
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
		os.WriteFile(filepath.Join(configDir, configTemplate), []byte(input), 0777),
		os.WriteFile(filepath.Join(configDir, configContextFile), configContextBytes, 0777),
		os.WriteFile(projectContextFile, projectContextBytes, 0777),
	)
	if err != nil {
		t.Fatal(err.Error())
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
		t.Error(err.Error())
	}
	result := string(resultBytes)
	exp := `# Hello, Potato at cd/ef.
# Things:
# • a, A
# • b, B
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
		t.Fatal(err.Error())
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
		t.Error(err.Error())
	}
	result = string(resultBytes)
	exp = `# Hello, Potato at cd/ef.
# Things:
# • a, A
# • b, B
# • c, C
`
	if result != exp {
		t.Errorf("wrong results: %v", result)
	}
	ok, err = Run(true)
	if !ok || err != nil {
		t.Errorf("wrong result: %v %v", ok, err)
	}
}
