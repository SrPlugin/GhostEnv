package main

import (
	"github.com/SrPlugin/GhostEnv/internal/vault"
)

func getVaultService(environment string) (vault.Service, error) {
	vaultPath, _, err := vault.GetVaultPath(environment)
	if err != nil {
		return nil, err
	}
	return vault.NewService(vaultPath), nil
}
