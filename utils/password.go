package utils

import (
	"errors"
	"github.com/manifoldco/promptui"
	"unicode"
)

func PasswordPrompt() string {
	prompt := promptui.Prompt{
		Label:    "Password",
		Validate: validate,
		Mask:     ' ',
	}
	password, err := prompt.Run()
	if err != nil {
		log.Fatal(err)
	}
	return password
}

func PrivateKeyPrompt() string {
	prompt := promptui.Prompt{
		Label:    "🔑 Private Key",
		Validate: validatePrivateKey,
		Mask:     ' ',
	}
	privateKey, err := prompt.Run()
	if err != nil {
		log.Fatal(err)
	}
	return privateKey
}

func validate(input string) error {
	if input == "" || !strongPassword(input) {
		return errors.New("enter a valid password")
	}
	return nil
}

func validatePrivateKey(input string) error {
	if input == "" {
		return errors.New("enter a valid private key")
	}
	return nil
}

func AssignPassword() string {
	return PasswordPrompt()
}

//This function checks if the password is strong enough or not
func strongPassword(input string) bool {
	l, u, p, d := 0, 0, 0, 0
	if len(input) >= 8 {
		for _, char := range input {
			switch {
			case unicode.IsUpper(char):
				u += 1
			case unicode.IsLower(char):
				l += 1
			case unicode.IsNumber(char):
				d += 1
			case unicode.IsPunct(char) || unicode.IsSymbol(char):
				p += 1
			}
		}
	}
	return (l >= 1 && u >= 1 && p >= 1 && d >= 1) && (l+u+p+d == len(input))
}
