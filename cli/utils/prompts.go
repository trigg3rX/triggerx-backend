package utils

import (
	"fmt"
	"os"
	"strings"
	"regexp"

	"golang.org/x/term"
	"github.com/trigg3rX/triggerx-keeper/pkg/core/errors"
)

// StringPrompt prompts the user for a string input.
func StringPrompt(label string, allowEmpty bool, hasOptions bool, options []string) (string, error) {
	var input string
	for {
		fmt.Fprint(os.Stderr, label)
		if _, err := fmt.Fscanln(os.Stdin, &input); err != nil {
			return "", err
		}

		input = strings.TrimSpace(input)

		if !allowEmpty && len(input) == 0 {
			fmt.Fprintln(os.Stderr, "Input cannot be empty")
			continue
		}

		if hasOptions {
			validOption := false
			for _, opt := range options {
				if strings.EqualFold(input, opt) {
					validOption = true
					break
				}
			}
			if !validOption {
				fmt.Fprintf(os.Stderr, "Invalid input. Must be one of: %s\n", strings.Join(options, ", "))
				continue
			}
		}

		break
	}
	return input, nil
}

// IntegerPrompt prompts the user for an integer input.
func IntegerPrompt(label string) (int, error) {
	var input int
	for {
		fmt.Fprint(os.Stderr, label)
		if _, err := fmt.Fscanln(os.Stdin, &input); err != nil {
			return 0, err
		}
		if input > 0 {
			break
		}
		fmt.Fprintln(os.Stderr, "Input must be a positive integer")
	}
	return input, nil
}

func PasswordPrompt(label string, toConfirm bool) (string, error) {
	var password string
	for {
		fmt.Fprint(os.Stderr, label)
		bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return "", err
		}
		password = string(bytePassword)
		fmt.Fprintln(os.Stderr)

		if err := validatePassword(password); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			continue
		}

		if toConfirm {
			fmt.Fprint(os.Stderr, "Confirm password: ")
			byteConfirmPassword, err := term.ReadPassword(int(os.Stdin.Fd()))
			if err != nil {
				return "", err
			}
			confirmPassword := string(byteConfirmPassword)
			fmt.Fprintln(os.Stderr)

			if password != confirmPassword {
				fmt.Fprintln(os.Stderr, "Passwords do not match. Please try again.")
				continue
			}
		}

		break
	}
	return password, nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("%w: password must be at least 8 characters long", errors.ErrInvalidPassword)
	}
	
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()]`).MatchString(password)
	
	if !hasUpper || !hasLower || !hasDigit || !hasSpecial {
		return fmt.Errorf("%w: password must contain uppercase, lowercase, number, and special character", errors.ErrInvalidPassword)
	}
	
	return nil
}

// ConfirmPrompt prompts the user for a confirmation input.
func ConfirmPrompt(label string) (bool, error) {
	var input string
	for {
		fmt.Fprint(os.Stderr, label)
		if _, err := fmt.Fscanln(os.Stdin, &input); err != nil {
			return false, err
		}
		switch strings.ToLower(strings.Trim(input, " ")) {
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		default:
			fmt.Fprintln(os.Stderr, "Invalid input. Please enter 'y' or 'n'")
		}
	}
}
