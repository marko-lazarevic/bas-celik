// Package localization provides localization formatting functions.
package localization

import "strings"

// FormatYesNo returns the localized string for a boolean value
func FormatYesNo(a bool, script Language) string {
	switch script {
	case SrLatin:
		if a {
			return "Da"
		}
		return "Ne"
	case SrCyrillic:
		if a {
			return "Да"
		}
		return "Не"
	default:
		if a {
			return "Yes"
		}
		return "No"
	}
}

// FormatDate expects a pointer to a date in the format DDMMYYYY.
// Modifies, in place, date to format DD.MM.YYYY.
func FormatDate(in *string) {
	chars := strings.Split(*in, "")
	if len(chars) != 8 {
		return
	}
	if chars[4] == "0" {
		*in = "Nije dostupan"
		return
	}
	*in = chars[0] + chars[1] + "." + chars[2] + chars[3] + "." + chars[4] + chars[5] + chars[6] + chars[7] + "."
}

// FormatDateYMD expects a pointer to a date in the format YYYYMMDD.
// Modifies, in place, date to format DD.MM.YYYY.
func FormatDateYMD(in *string) {
	chars := strings.Split(*in, "")
	if len(chars) != 8 {
		return
	}
	*in = chars[6] + chars[7] + "." + chars[4] + chars[5] + "." + chars[0] + chars[1] + chars[2] + chars[3]
}

// JoinWithComma joins a list of strings into a single string
// separating them with a comma and a space.
// Empty strings are skipped.
func JoinWithComma(strs ...string) string {
	var nonemptyStrings []string
	for _, str := range strs {
		if str != "" {
			nonemptyStrings = append(nonemptyStrings, str)
		}
	}

	return strings.Join(nonemptyStrings, ", ")
}
