package utils

import (
	"errors"
	"fmt"
	"github.com/bilisound/server/internal/config"
	"github.com/bilisound/server/internal/structure"
	"github.com/buger/jsonparser"
	"github.com/go-resty/resty/v2"
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

func getVideo(html string) (string, string, error) {
	initialStateRegex := regexp.MustCompile(`window\.__INITIAL_STATE__=(\{.+});`)
	playInfoRegex := regexp.MustCompile(`window\.__playinfo__=(\{.+})<\/script><script>`)

	initialState, err := ExtractContent(initialStateRegex, html, ExtractJSONOptions{})
	if err != nil {
		return "", "", err
	}

	playInfo, err := ExtractContent(playInfoRegex, html, ExtractJSONOptions{})
	if err != nil {
		return "", "", err
	}

	return initialState, playInfo, nil
}

func ParseVideo() error {
	// Create a Resty Client
	client := resty.New()

	resp, err := client.R().
		EnableTrace().
		SetHeader("User-Agent", config.Global.MustString("request.userAgent")).
		Get("https://www.bilibili.com/video/BV16D4y1b7P7/")

	if err != nil {
		return err
	}

	err = ParseHTML(resp.String())
	return err
}

func ParseHTML(html string) error {
	initialState, _, err := getVideo(html)
	if err != nil {
		return err
	}

	//bvid, err := jsonparser.GetString([]byte(initialState), "videoData", "bvid")
	//fmt.Println("bvid:", bvid)

	video := structure.Video{}

	paths := [][]string{
		{"videoData", "bvid"},
		{"videoData", "aid"},
		{"videoData", "title"},
	}
	jsonparser.EachKey([]byte(initialState), func(idx int, value []byte, vt jsonparser.ValueType, err error) {
		switch idx {
		case 0:
			video.Bvid, err = jsonparser.ParseString(value)
			break
		case 1:
			video.Aid, err = jsonparser.ParseInt(value)
			break
		case 2:
			video.Title, err = jsonparser.ParseString(value)
			break
		}
	}, paths...)

	fmt.Print(video)
	return nil
}
