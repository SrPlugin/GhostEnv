package shamir

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lafriks/go-shamir"
)

func SplitSecret(secret []byte, parts, threshold int) ([][]byte, error) {
	return shamir.Split(secret, parts, threshold)
}

func CombineShares(shares [][]byte) ([]byte, error) {
	if len(shares) < 2 {
		return nil, fmt.Errorf("at least 2 shares required")
	}
	shareSlice := make([][]byte, len(shares))
	for i, s := range shares {
		shareSlice[i] = s
	}
	return shamir.Combine(shareSlice...)
}

func WriteShareToFile(share []byte, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	encoded := base64.StdEncoding.EncodeToString(share)
	return os.WriteFile(path, []byte(encoded), 0600)
}

func ReadShareFromFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return base64.StdEncoding.DecodeString(string(data))
}

func zeroBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}

func ZeroShares(shares [][]byte) {
	for _, s := range shares {
		zeroBytes(s)
	}
}
