package utils

import (
	"os"
	"path/filepath"

	"github.com/btcsuite/btcd/btcutil"
)

func ReadCertFile(dir, filename string) ([]byte, error) {
	certPath := filepath.Join(btcutil.AppDataDir(dir, false), filename)
	cert, err := os.ReadFile(certPath)
	if err != nil {
		return nil, err
	}

	return cert, nil
}
