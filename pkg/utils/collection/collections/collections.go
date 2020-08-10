/**
 *
 * @author yinjk
 * @create 2019-06-14 10:36
 */
package collections

import "strings"

func IsIn(key interface{}, values ...interface{}) bool {
	for _, value := range values {
		if key == value {
			return true
		}
	}
	return false
}

func IsStringIn(key string, values ...string) bool {
	for _, value := range values {
		if key == value {
			return true
		}
	}
	return false
}

func MapContains(maps map[string]string, key string) bool {
	for k := range maps {
		if k == key {
			return true
		}
	}
	return false
}

func IsInIgnoreCase(key string, values string) bool {
	if values != "" {
		s := strings.Split(values, ",")
		for _, value := range s {
			if strings.ToLower(key) == strings.ToLower(value) {
				return true
			}
		}
	}

	return false
}
