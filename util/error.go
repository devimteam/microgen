package util

import (
	"fmt"
	"strings"
)

func ComposeErrors(errs []error) error {
	if len(errs) > 0 {
		var strs []string
		for _, err := range errs {
			if err != nil {
				strs = append(strs, err.Error())
			}
		}
		if len(strs) == 1 {
			return fmt.Errorf(strs[0])
		}
		if len(strs) > 0 {
			return fmt.Errorf("many errors:\n%v", strings.Join(strs, "\n"))
		}
	}
	return nil
}
