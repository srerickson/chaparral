package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const projectDir = "." + chaparral

var (
	ErrNotProject = errors.New("directory context does not reference a remote object")
)

type Project struct {
	// repo/store_path
	StorageRootID string `json:"storage_root_id,omitempty"`
	ObjectID      string `json:"object_id"`
	Version       int    `json:"version"`

	path string // path to the project
}

func (p *Project) Save() error {
	if p.path == "" {
		return errors.New("project not initialized")
	}
	byts, err := json.Marshal(p)
	if err != nil {
		return err
	}
	conf := filepath.Join(p.path, projectDir, configFile)
	return os.WriteFile(conf, byts, fileMode)
}

func InitProject(dir string, p Project) error {
	newDir := filepath.Join(dir, projectDir)
	if err := os.Mkdir(newDir, dirMode); err != nil {
		return err
	}
	p.path = dir
	return p.Save()
}

func GetProject(dir string) (p Project, err error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return
	}
	byts, err := os.ReadFile(filepath.Join(abs, projectDir, configFile))
	if err != nil {
		if os.IsNotExist(err) {
			err = errors.Join(err, fmt.Errorf("%s: %w", abs, ErrNotProject))
		}
		return
	}

	if err = json.Unmarshal(byts, &p); err != nil {
		return
	}
	p.path = abs
	return
}

// Check if dir or any of it's parent directories is a project, and return it.
// If dir is empty, search the working directory.
func SearchProject(dir string) (p Project, err error) {
	if dir == "" || dir == "." {
		dir, err = os.Getwd()
		if err != nil {
			return
		}
	}
	for {
		p, err = GetProject(dir)
		if err != nil && errors.Is(err, ErrNotProject) {
			parent := filepath.Dir(dir)
			if dir == parent {
				return
			}
			dir = parent
			continue
		}
		if err != nil {
			return
		}
		break
	}
	return
}

func (p Project) Path() string {
	return p.path
}

func (p Project) Empty() bool {
	return p.path == ""
}
