package importer

import "strings"

func matchTagsAgainstCompulsoryTagList(tags map[string]string, tagList []string) bool {
	for _, name := range tagList {

		feature := strings.Split(name, "~")
		foundVal, foundKey := tags[feature[0]]

		// key check
		if !foundKey {
			return false
		}

		// value check
		if len(feature) > 1 {
			if foundVal != feature[1] {
				return false
			}
		}
	}

	return true
}

// check tags contain features from a groups of whitelists
func containsValidTags(tags map[string]string, group map[string][]string) bool {
	for _, list := range group {
		if matchTagsAgainstCompulsoryTagList(tags, list) {
			return true
		}
	}
	return false
}

// trim leading/trailing spaces from keys and values
func trimTags(tags map[string]string) map[string]string {
	trimmed := make(map[string]string)
	for k, v := range tags {
		trimmed[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	return trimmed
}

// check if a tag list is empty or not
func hasTags(tags map[string]string) bool {
	if n := len(tags); n == 0 {
		return false
	}
	return true
}

func BuildTags(tagList string) map[string][]string {
	conditions := make(map[string][]string)
	for _, group := range strings.Split(tagList, ",") {
		conditions[group] = strings.Split(group, "+")
	}
	return conditions
}
