package utils

import (
	"errors"
	"regexp"
)

type ExtractJSONOptions struct {
	ParsePrefix string
	ParseSuffix string
	Index       int
}

func ExtractContent(regex *regexp.Regexp, str string, options ExtractJSONOptions) (string, error) {
	parsePrefix := options.ParsePrefix
	parseSuffix := options.ParseSuffix
	index := options.Index

	if index == 0 {
		index = 1
	}

	matches := regex.FindStringSubmatch(str)
	if len(matches) <= index {
		return "", errors.New("unable to extract contents from provided string")
	}

	jsonStr := parsePrefix + matches[index] + parseSuffix

	return jsonStr, nil
}
