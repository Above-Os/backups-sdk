package utils

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"syscall"

	"golang.org/x/term"
)

func InputPasswordWithConfirm(confirmRequired bool) (string, error) {
	if confirmRequired {
		fmt.Println("\nPlease create a password for this backup. This password will be required to restore your data in the future. The system will NOT save or store this password, so make sure to remember it. If you lose or forget this password, you will not be able to recover your backup.")
	}

	var password []byte
	var confirmed []byte
	_ = password

	for {
		fmt.Print("\nEnter password for repository: ")
		password, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			log.Fatalf("Failed to read password: %v", err)
			return "", err
		}
		password = bytes.TrimSpace(password)
		if len(password) == 0 {
			continue
		}
		confirmed = password
		if !confirmRequired {
			break
		}
		fmt.Print("\nRe-enter the password to confirm: ")
		confirmed, err = term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			log.Fatalf("Failed to read re-enter password: %v", err)
			return "", err
		}
		if !bytes.Equal(password, confirmed) {
			fmt.Printf("\nPasswords do not match. Please try again.\n")
			continue
		}

		break
	}
	fmt.Printf("\n\n")

	return string(confirmed), nil
}

func ContainsPathSeparator(s string) bool {
	return strings.ContainsAny(s, `/\`)
}
