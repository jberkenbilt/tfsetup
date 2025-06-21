package tfsetup

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Masterminds/sprig/v3"
	"github.com/jberkenbilt/tfsetup/util"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"text/template"
)

const projectContextFile = "tfsetup-context.json"
const configPath = "tfsetup-config"
const configContextFile = "context.json"
const configTemplate = "setup.tmpl"
const SetupFile = "setup.tf"

var tfCommand = func() string {
	_, err := exec.LookPath("tofu")
	if err == nil {
		return "tofu"
	}
	_, err = exec.LookPath("terraform1")
	if err == nil {
		return "terraform"
	}
	return ""
}()

type templateContext struct {
	Config        any
	Project       any
	Path          string
	relConfigPath string // not part of context since private
}

func newTemplateContext(
	projectContextBytes []byte,
	configContextBytes []byte,
	relPath string,
	relConfigPath string,
) (*templateContext, error) {
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
	return &templateContext{
		Config:        configContext,
		Project:       projectContext,
		Path:          relPath,
		relConfigPath: relConfigPath,
	}, nil
}

func (tc *templateContext) renderFile(inFile, outFile string, allowChanges bool) (bool, error) {
	input, err := os.ReadFile(inFile)
	if err != nil {
		return false, fmt.Errorf("reading %s: %w", inFile, err)
	}
	rendered, err := tc.render(inFile, input)
	if err != nil {
		return false, err
	}
	if tfCommand != "" && strings.HasSuffix(outFile, ".tf") {
		cmd := exec.Command(tfCommand, "fmt", "-")
		cmd.Stdin = bytes.NewReader(rendered)
		var out bytes.Buffer
		cmd.Stdout = &out
		var errOut bytes.Buffer
		cmd.Stderr = &errOut
		err = cmd.Run()
		if err != nil {
			out.Reset()
			out.Write(rendered)
			_, _ = fmt.Fprintf(&out, "/* --- OUTPUT FROM %s fmt ---\n%s\n*/\n", tfCommand, errOut.String())
			_, _ = fmt.Fprintf(os.Stderr, "WARNING: %s fmt: failed; output appended to file\n", tfCommand)
		}
		rendered = out.Bytes()
	}
	// It's okay if the output file doesn't exist at this point.
	origOutputBytes, _ := os.ReadFile(outFile)
	if slices.Equal(origOutputBytes, rendered) {
		return true, nil
	}
	if !allowChanges {
		return false, nil
	}
	_ = os.Remove(outFile)
	err = os.WriteFile(outFile, rendered, 0444)
	if err != nil {
		return false, fmt.Errorf("error writing %s: %w", outFile, err)
	}
	fmt.Printf("updated %s\n", outFile)
	return false, nil
}

func (tc *templateContext) render(name string, input []byte) ([]byte, error) {
	tpl, err := template.New(name).Funcs(sprig.FuncMap()).Parse(string(input))
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}
	var buf bytes.Buffer
	err = tpl.Execute(&buf, tc)
	if err != nil {
		return nil, fmt.Errorf("evaluate template: %w", err)
	}
	return buf.Bytes(), nil
}

func makeContext() (*templateContext, error) {
	projectContextBytes, err := os.ReadFile(projectContextFile)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", projectContextFile, err)
	}
	relConfigPath, relPath, err := util.FindDir(configPath)
	if err != nil {
		return nil, fmt.Errorf("unable to find %s above current directory: %w", relConfigPath, err)
	}
	configContextBytes, err := os.ReadFile(filepath.Join(relConfigPath, configContextFile))
	if err != nil {
		return nil, fmt.Errorf("reading %s from %s: %w", configContextFile, relConfigPath, err)
	}
	return newTemplateContext(projectContextBytes, configContextBytes, relPath, relConfigPath)
}

// Run checks or updates the setup file. The boolean return value indicates
// whether the file was already up-to-date.
func Run(allowChanges bool) (bool, error) {
	tpl, err := makeContext()
	if err != nil {
		return false, err
	}
	allOkay := true
	var allErrors []error
	handle := func(inFile, outFile string) {
		ok, err := tpl.renderFile(inFile, outFile, allowChanges)
		if err != nil {
			allErrors = append(allErrors, err)
		}
		if !ok {
			allOkay = false
		}
	}
	handle(filepath.Join(tpl.relConfigPath, configTemplate), SetupFile)
	entries, err := os.ReadDir(".")
	if err != nil {
		return false, fmt.Errorf("error reading current directory: %w", err)
	}
	for _, entry := range entries {
		name := entry.Name()
		outFile := strings.TrimSuffix(name, ".tfsetup.tmpl")
		if outFile != name {
			handle(name, outFile)
		}
	}
	return allOkay, errors.Join(allErrors...)
}

func Render() error {
	tpl, err := makeContext()
	if err != nil {
		return err
	}
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("reading standard input: %w", err)
	}
	output, err := tpl.render("stdin", input)
	if err != nil {
		return err
	}
	_, _ = os.Stdout.Write(output)
	return nil
}
