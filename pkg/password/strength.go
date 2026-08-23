package password

import "unicode"

func Strength(value string) string {
	score := 0
	if len([]rune(value)) >= 12 {
		score++
	}
	if len([]rune(value)) >= 16 {
		score++
	}
	hasUpper, hasLower, hasNumber, hasSymbol := false, false, false, false
	for _, r := range value {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasNumber = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSymbol = true
		}
	}
	if value == "" || value == "123456" || value == "password" || value == "123456789" {
		return "weak"
	}
	if hasUpper {
		score++
	}
	if hasLower {
		score++
	}
	if hasNumber {
		score++
	}
	if hasSymbol {
		score++
	}
	if score >= 5 {
		return "strong"
	}
	if score >= 3 {
		return "medium"
	}
	return "weak"
}
