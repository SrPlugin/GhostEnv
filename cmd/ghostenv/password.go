package main

import (
	"fmt"
	"os"
	"unsafe"

	"golang.org/x/term"
)

func zeroBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}

func getPassword(masterPassword string) (string, error) {
	if masterPassword != "" {
		return masterPassword, nil
	}

	fmt.Print("Enter Master Password: ")
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}
	if len(bytePassword) == 0 {
		return "", fmt.Errorf("password cannot be empty")
	}

	password := string(bytePassword)
	zeroBytes(bytePassword)

	ptr := (*unsafe.Pointer)(unsafe.Pointer(&bytePassword))
	*ptr = nil

	return password, nil
}
