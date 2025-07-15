package utils

import (
	"github.com/nyaruka/phonenumbers"
	"strings"
)

func RemoveDuplicates(numbers []string) []string {
	uniqueNumbers := make(map[string]bool)
	var result []string

	for _, number := range numbers {
		if _, exists := uniqueNumbers[number]; !exists {
			uniqueNumbers[number] = true
			result = append(result, number)
		}
	}
	return result
}

// NormalizePhoneNumber uses nyaruka/phonenumbers to standardize
func NormalizePhoneNumber(raw string, defaultRegion string) (string, error) {
	num, err := phonenumbers.Parse(raw, defaultRegion)
	if err != nil || !phonenumbers.IsValidNumber(num) {
		return "", err
	}
	// E.164 is the +256702xxxxxx format
	return phonenumbers.Format(num, phonenumbers.E164), nil
}

// ExtractUniquePhones extracts unique phone numbers from a map using specified keys
func ExtractUniquePhones(data map[string]interface{}, keys []string, defaultRegion string) []string {
	seen := make(map[string]struct{})
	result := []string{}
	for _, key := range keys {
		if val, ok := data[key]; ok {
			if str, ok := val.(string); ok {
				num, err := NormalizePhoneNumber(str, defaultRegion)
				num = strings.ReplaceAll(num, "+", "")
				if err == nil && num != "" {
					if _, exists := seen[num]; !exists {
						seen[num] = struct{}{}
						result = append(result, num)
					}
				}
			}
		}
	}
	return result
}
