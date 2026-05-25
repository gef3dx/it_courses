package slug

import "strings"

func Make(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))

	replacer := strings.NewReplacer(
		"а", "a", "б", "b", "в", "v", "г", "g", "д", "d",
		"е", "e", "ё", "e", "ж", "zh", "з", "z", "и", "i",
		"й", "y", "к", "k", "л", "l", "м", "m", "н", "n",
		"о", "o", "п", "p", "р", "r", "с", "s", "т", "t",
		"у", "u", "ф", "f", "х", "h", "ц", "ts", "ч", "ch",
		"ш", "sh", "щ", "sch", "ъ", "", "ы", "y", "ь", "",
		"э", "e", "ю", "yu", "я", "ya",
	)
	value = replacer.Replace(value)

	var builder strings.Builder
	lastDash := false
	for _, ch := range value {
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') {
			builder.WriteRune(ch)
			lastDash = false
			continue
		}
		if !lastDash {
			builder.WriteByte('-')
			lastDash = true
		}
	}

	result := strings.Trim(builder.String(), "-")
	if result == "" {
		return "item"
	}

	return result
}
