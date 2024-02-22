package testutil

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/srerickson/chaparral"
	"github.com/srerickson/ocfl-go"
)

func UploadDir(cli *chaparral.Client, uper *chaparral.Uploader, dir string, alg string) (ocfl.PathMap, error) {
	fsys := os.DirFS(dir)
	ctx := context.Background()
	paths := ocfl.PathMap{}
	walkFn := func(name string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if info.Type().IsDir() {
			return nil
		}
		if !info.Type().IsRegular() {
			return errors.New("irregular file :" + name)
		}
		var f fs.File
		f, err = fsys.Open(name)
		if err != nil {
			return err
		}
		defer func() {
			if closeErr := f.Close(); closeErr != nil {
				err = errors.Join(err, closeErr)
			}
		}()
		digester := ocfl.NewDigester(alg)
		upResult, err := cli.Upload(ctx, uper.UploadPath, io.TeeReader(f, digester))
		if err != nil {
			return err
		}
		upDigest, ok := upResult.Digests[alg]
		if !ok || upDigest == "" {
			return fmt.Errorf("upload result doesn't included %s as expected", alg)
		}
		if digester.String() != upDigest {
			return fmt.Errorf("unexpected %s for %s from uploader", alg, name)
		}
		paths[name] = upDigest
		return nil
	}
	if err := fs.WalkDir(fsys, ".", walkFn); err != nil {
		return nil, err
	}
	return paths, nil
}
