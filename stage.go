package chaparral

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/srerickson/ocfl-go"
)

var (
	DIGEST_CONCURRENCY = runtime.NumCPU()
)

var dotFileRE = regexp.MustCompile(`^(.*/)?\.[^\/]+$`)

type Stage struct {
	// Alg is the digest algorithm used to generate digests in State and Content.
	Alg string
	// Logical Path to digest
	State map[string]string
	// Content
	Content map[string][]string

	// Regexp for directories to skip during Walk
	ingnoreREs []*regexp.Regexp
	onRead     func(int, error)
}

type StageOption func(*Stage) error

func NewStage(alg string, opts ...StageOption) (*Stage, error) {
	stage := &Stage{
		Alg:     alg,
		State:   map[string]string{},
		Content: map[string][]string{},
	}
	for _, opt := range opts {
		if err := opt(stage); err != nil {
			return nil, err
		}
	}
	return stage, nil
}

// OnRead is used to set a function that will be called for each Read() during
// the staging process. This is used to make ui that track progress of the
// staging process.
func OnRead(fn func(int, error)) StageOption {
	return func(stage *Stage) error {
		stage.onRead = fn
		return nil
	}
}

func StageSkipDirRE(re *regexp.Regexp) StageOption {
	return func(stage *Stage) error {
		stage.ingnoreREs = append(stage.ingnoreREs, re)
		return nil
	}
}

func StageSkipHidden() StageOption {
	return func(stage *Stage) error {
		stage.ingnoreREs = append(stage.ingnoreREs, dotFileRE)
		return nil
	}
}

func AddDir(dir string) StageOption {
	return func(stage *Stage) error {
		ctx := context.Background()
		dir = filepath.Clean(dir)
		fsys := os.DirFS(dir)
		if stage.onRead != nil {
			fsys = &meterFS{base: fsys, onRead: stage.onRead}
		}
		var walkErr error
		setupFn := func(add func(string, ...string) bool) {
			walkErr = fs.WalkDir(fsys, ".", func(name string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() {
					for _, re := range stage.ingnoreREs {
						if re.MatchString(name) {
							return fs.SkipDir
						}
					}
					return nil
				}
				if !d.Type().IsRegular() {
					// todo -- log that a file is ignored but continue?
					return errors.New("irregular file " + name)
				}
				for _, re := range stage.ingnoreREs {
					if re.MatchString(name) {
						return nil
					}
				}
				if !add(name, stage.Alg) {
					return errors.New("digest canceled")
				}
				return nil
			})
		}
		result := func(name string, digs ocfl.DigestSet, err error) error {
			if err != nil {
				return err
			}
			dig := digs[stage.Alg]
			realName := filepath.Join(dir, filepath.FromSlash(name))
			stage.State[name] = dig
			stage.Content[dig] = append(stage.Content[dig], realName)

			return nil
		}
		digestErr := ocfl.DigestFS(ctx, ocfl.NewFS(fsys), setupFn, result)
		if err := errors.Join(digestErr, walkErr); err != nil {
			return err
		}
		return nil
	}
}

type meterFS struct {
	base   fs.FS
	onRead func(int, error)
}

func (mfs *meterFS) Open(name string) (fs.File, error) {
	f, err := mfs.base.Open(name)
	if err != nil {
		return nil, err
	}
	return &meterFile{File: f, onRead: mfs.onRead}, nil
}

func (mfs *meterFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return fs.ReadDir(mfs.base, name)
}

type meterFile struct {
	fs.File
	onRead func(int, error)
}

func (mf *meterFile) Read(p []byte) (int, error) {
	s, err := mf.File.Read(p)
	if mf.onRead != nil {
		mf.onRead(s, err)
	}
	return s, err
}
