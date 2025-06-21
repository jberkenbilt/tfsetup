package util_test

import (
	"github.com/jberkenbilt/tfsetup/util"
	"os"
	"path/filepath"
	"testing"
)

func TestFindDir(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err.Error())
	}
	expTarget := filepath.Join(filepath.Dir(cwd), ".git")
	target, rel, err := util.FindDir(".git")
	if err != nil {
		t.Error(err.Error())
	}
	if target != expTarget || rel != "util" {
		t.Errorf("%v, %v", target, expTarget)
	}
	_, _, err = util.FindDir("does-not-exist")
	if err.Error() != "does-not-exist not found as directory at or above current directory" {
		t.Errorf("wrong error: %v", err)
	}
	_, _, err = util.FindDir("go.mod")
	if err.Error() != "go.mod not found as directory at or above current directory" {
		t.Errorf("wrong error: %v", err)
	}
}
