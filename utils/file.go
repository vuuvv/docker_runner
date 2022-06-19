package utils

import (
	"encoding/base64"
	"github.com/vuuvv/errors"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"path/filepath"
)

func MountSecret(path string, secret string, perm os.FileMode) (err error) {
	zap.L().Info("MountSecret", zap.String("path", path), zap.String("secret", secret))
	if _, err = os.Stat(path); err == nil {
		zap.L().Info("MountSecret remove path", zap.String("path", path))
		err = os.RemoveAll(path)
		if err != nil {
			return errors.Errorf("Mount secret [%s] error, cannot remove old file: %s", path, err.Error())
		}
	}

	if secret == "" {
		return nil
	}

	dir := filepath.Dir(path)
	zap.L().Info("MountSecret create directory", zap.String("dir", dir))
	if err = os.MkdirAll(filepath.Dir(dir), os.ModeDir|0755); err != nil {
		return errors.Errorf("Mount secret [%s] error: %s", path, err.Error())
	}

	bs, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return errors.Errorf("Mount secret [%s] error, secret should be valid base64 string: %s", path, err.Error())
	}

	zap.L().Info("MountSecret write file", zap.String("bs", string(bs)))
	err = ioutil.WriteFile(path, bs, os.ModeDir|perm)
	if err != nil {
		return errors.Errorf("Mount secret [%s] error, cannot write to file: %s", path, err.Error())
	}
	return nil
}
