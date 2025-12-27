// Package translation handles loading and retrieving translations for different languages.
package translation

import (
	"embed"
	"encoding/json"
	"fmt"

	"github.com/ubavic/bas-celik/v2/localization"
)

var translations map[localization.Language]map[string]string

var currentLanguage localization.Language

// SetTranslations loads translations from the embedded filesystem.
func SetTranslations(embedFS embed.FS) error {
	translations = make(map[localization.Language]map[string]string)

	languages := []localization.Language{localization.SrLatin, localization.SrCyrillic, localization.En}
	for _, lang := range languages {
		langJSON, err := embedFS.ReadFile("embed/translation/" + string(lang) + ".json")
		if err != nil {
			return fmt.Errorf("reading %s translation: %w", lang, err)
		}

		langMap := make(map[string]string)

		err = json.Unmarshal(langJSON, &langMap)
		if err != nil {
			return fmt.Errorf("%s translation unmarshal: %w", lang, err)
		}

		translations[lang] = langMap
	}

	return nil
}

// SetLanguage sets the current language for translations.
func SetLanguage(lang int) {
	switch lang {
	case 2:
		currentLanguage = localization.En
	case 1:
		currentLanguage = localization.SrCyrillic
	default:
		currentLanguage = localization.SrLatin
	}
}

// CurrentLanguage returns the currently set language.
func CurrentLanguage() localization.Language {
	return currentLanguage
}

// Translate returns the translation for the given ID in the current language.
func Translate(id string, vals ...any) string {
	translation := translations[currentLanguage][id]

	if len(vals) == 0 {
		return translation
	}
	return fmt.Sprintf(translation, vals...)
}

// EnglishTranslation returns the English translation for the given ID.
func EnglishTranslation(id string) string {
	return translations[localization.En][id]
}
