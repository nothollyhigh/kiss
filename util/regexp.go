package util

import (
	"regexp"
)

func IsChineseName(str string) bool {
	ok, err := regexp.MatchString(`^([\u4e00-\u9fa5][·\u4e00-\u9fa5]{0,30}[\u4e00-\u9fa5])$`, str)
	return ok && err == nil
}

func IsEmail(str string) bool {
	ok, err := regexp.MatchString(`[\w+\.]+[\w+]+@+[0-9A-Za-z]+\.+[A-Za-z]+$`, str)
	return ok && err == nil
}

func IsMobilePhone(str string) bool {
	ok, err := regexp.MatchString(`^((\+86)|(86))?(-|\s)?1\d{10}$`, str)
	return ok && err == nil
}

func IsOnlyChinese(str string) bool {
	var hzRegexp = regexp.MustCompile("^[\u4e00-\u9fa5]+$")
	return hzRegexp.MatchString(str)
}
