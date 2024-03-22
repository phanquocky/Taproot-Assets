package utils

import (
	"log"
	"os"
	"path/filepath"

	"github.com/btcsuite/btcd/btcutil"
)

func ReadCertFile() ([]byte, error) {
	certPath := filepath.Join(btcutil.AppDataDir("btcwallet", false), "rpc.cert")
	cert, err := os.ReadFile(certPath)
	if err != nil {
		log.Println("cannot read cert file, ", err)
		return nil, err
	}

	return cert, nil
}
