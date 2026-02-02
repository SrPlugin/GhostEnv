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

func getPassword(flagValue string) ([]byte, error) {
	if env := os.Getenv("GHOSTENV_PASS"); env != "" {
		return []byte(env), nil
	}
	if flagValue != "" {
		return []byte(flagValue), nil
	}

	fmt.Print("Enter Master Password: ")
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return nil, fmt.Errorf("failed to read password: %w", err)
	}
	if len(bytePassword) == 0 {
		return nil, fmt.Errorf("password cannot be empty")
	}

	result := make([]byte, len(bytePassword))
	copy(result, bytePassword)
	zeroBytes(bytePassword)
	ptr := (*unsafe.Pointer)(unsafe.Pointer(&bytePassword))
	*ptr = nil

	return result, nil
}

func getNewPassword() ([]byte, error) {
	fmt.Print("Enter new password: ")
	newPw, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return nil, fmt.Errorf("failed to read password: %w", err)
	}
	if len(newPw) == 0 {
		return nil, fmt.Errorf("new password cannot be empty")
	}

	fmt.Print("Confirm new password: ")
	confirmPw, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		zeroBytes(newPw)
		return nil, fmt.Errorf("failed to read password: %w", err)
	}
	if string(newPw) != string(confirmPw) {
		zeroBytes(newPw)
		zeroBytes(confirmPw)
		return nil, fmt.Errorf("passwords do not match")
	}

	result := make([]byte, len(newPw))
	copy(result, newPw)
	zeroBytes(newPw)
	zeroBytes(confirmPw)
	ptrNew := (*unsafe.Pointer)(unsafe.Pointer(&newPw))
	*ptrNew = nil
	ptrConfirm := (*unsafe.Pointer)(unsafe.Pointer(&confirmPw))
	*ptrConfirm = nil

	return result, nil
}
