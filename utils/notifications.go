package utils

import (
	"fmt"
	"strings"
)

// FilterRecipientAttributes filters recipient attributes based on consent value and ignore list
func FilterRecipientAttributes(attrs, ignoreAttrs []string, consentValue string) []string {
	if strings.ToLower(consentValue) == "no" {
		filtered := make([]string, 0, len(attrs))
		ignoreSet := make(map[string]struct{}, len(ignoreAttrs))
		for _, a := range ignoreAttrs {
			ignoreSet[strings.ToLower(a)] = struct{}{}
		}
		for _, attr := range attrs {
			if _, ignore := ignoreSet[strings.ToLower(attr)]; !ignore {
				filtered = append(filtered, attr)
			}
		}
		return filtered
	}
	return attrs
}

// DetectLanguage detects the language from payload using the configured attribute (default: "en")
func DetectLanguage(payload map[string]interface{}, langAttr string) string {
	lang := "en"
	if langAttr != "" {
		if l, ok := payload[langAttr]; ok && l != nil && fmt.Sprintf("%v", l) != "" {
			lang = fmt.Sprintf("%v", l)
		}
	}
	return lang
}

func MapKeys[M ~map[K]V, K comparable, V any](m M) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
