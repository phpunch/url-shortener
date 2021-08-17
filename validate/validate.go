package validate

import (
	"fmt"
	"regexp"
)

var blackLists = []string{"www.google.com"}

func CheckBlackList(url string) error {
	for _, blackList := range blackLists {
		matched, err := regexp.MatchString(
			fmt.Sprintf(`%s`, blackList),
			url,
		)
		if err != nil {
			return err
		}
		if matched {
			return fmt.Errorf("url is in blacklist")
		}
	}
	return nil
}
