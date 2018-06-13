package promicher

import (
	"fmt"
	"github.com/romana/rlog"
	"regexp"
)

func MergeDataMap(currentData map[string]string, newData map[string]string) map[string]string {
	res := make(map[string]string)

	for k, v := range newData {
		res[k] = v
	}

	for k, v := range currentData {
		res[k] = v
	}

	return res
}

func ApplyPattern(data, pattern string) (bool, string, error) {
	rgxp, err := regexp.Compile(pattern)
	if err != nil {
		return false, "", fmt.Errorf("bad pattern '%s': %s", pattern, err)
	}

	matches := rgxp.FindStringSubmatch(data)
	if len(matches) > 0 {
		res := matches[len(matches)-1]

		rlog.Debugf("'%s' MATCHED pattern '%s' => '%s'", data, pattern, res)

		return true, res, nil
	}

	rlog.Debugf("'%s' NOT MACHED pattern '%s'", data, pattern)

	return false, "", nil
}

func SelectData(data map[string]string, patterns []string) (map[string]string, error) {
	res := make(map[string]string)

	for k, v := range data {
		for _, pattern := range patterns {
			ok, newKey, err := ApplyPattern(k, pattern)
			if err != nil {
				return nil, fmt.Errorf("cannot apply pattern %s to key %s: %s", pattern, k, err)
			}

			if ok {
				res[newKey] = v
			}
		}
	}

	return res, nil
}
