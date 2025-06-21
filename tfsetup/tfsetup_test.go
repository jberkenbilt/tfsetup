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
const otherFile = `This is a text file.
{{- range .Project.things }}
• {{. | repeat 5}}
{{- end}}
`

func TestRender(t *testing.T) {
	configContextBytes := []byte(`{"name": "Potato"}`)
	projectContextBytes := []byte(`{"things": ["a", "b"]}`)
	tpl, err := newTemplateContext(projectContextBytes, configContextBytes, "x/y/z", "..")
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

func checkFile(t *testing.T, inFile string, exp string) {
	t.Helper()
	resultBytes, err := os.ReadFile(inFile)
	if err != nil {
		t.Error(err.Error())
		return
	}
	result := string(resultBytes)
	if result != exp {
		t.Errorf("wrong results: %v", result)
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
		os.WriteFile("a.txt.tfsetup.tmpl", []byte(otherFile), 0777),
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

	checkFile(
		t,
		SetupFile,
		`# Hello, Potato at cd/ef.
# Things:
# • a, A
# • b, B
`,
	)
	checkFile(
		t,
		"a.txt",
		`This is a text file.
• aaaaa
• bbbbb
`,
	)
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
	checkFile(
		t,
		SetupFile,
		`# Hello, Potato at cd/ef.
# Things:
# • a, A
# • b, B
# • c, C
`,
	)
	checkFile(
		t,
		"a.txt",
		`This is a text file.
• aaaaa
• bbbbb
• ccccc
`,
	)
	ok, err = Run(false)
	if !ok || err != nil {
		t.Errorf("wrong result: %v %v", ok, err)
	}
	ok, err = Run(true)
	if !ok || err != nil {
		t.Errorf("wrong result: %v %v", ok, err)
	}
}
