package config_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/srerickson/chaparral/cmd/chap/config"
)

func TestGetProject(t *testing.T) {
	tmp := t.TempDir()
	projectDir := filepath.Join(tmp, "project-dir")
	subDir := filepath.Join(projectDir, "a", "b", "c")
	be.NilErr(t, os.MkdirAll(subDir, 0777))
	// convert to path without symlinks, if necessary
	projectDir, err := filepath.EvalSymlinks(projectDir)
	be.NilErr(t, err)
	newProject := config.Project{
		StorageRootID: "root",
		ObjectID:      "object",
	}
	be.NilErr(t, config.InitProject(projectDir, newProject))
	// getting the project when working directory is outside fails
	_, err = config.SearchProject("")
	be.True(t, err != nil)
	be.True(t, errors.Is(err, config.ErrNotProject))
	// change working directory for test, then change back at the end.
	cwd, err := os.Getwd()
	be.NilErr(t, err)
	be.NilErr(t, os.Chdir(subDir))
	defer os.Chdir(cwd)
	p, err := config.SearchProject("")
	be.NilErr(t, err)
	// convert to path without symlinks, if necessary
	projectDir, err = filepath.EvalSymlinks(projectDir)
	be.NilErr(t, err)
	be.Equal(t, p.Path(), projectDir)
	be.Equal(t, p.ObjectID, newProject.ObjectID)
	be.Equal(t, p.StorageRootID, newProject.StorageRootID)
}
