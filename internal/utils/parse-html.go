package utils

import (
	"errors"
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
	playInfoRegex := regexp.MustCompile(`window\.__playinfo__=(\{.+})</script><script>`)

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

func ParseVideo(id string) (*structure.Video, error) {
	// Create a Resty Client
	client := resty.New()

	resp, err := client.R().
		EnableTrace().
		SetHeader("User-Agent", config.Global.MustString("request.userAgent")).
		Get("https://www.bilibili.com/video/" + id + "/")

	if err != nil {
		return nil, err
	}

	result, err := ParseHTMLRegular(resp.String())
	return result, err
}

func ParseHTMLRegular(html string) (*structure.Video, error) {
	initialState, _, err := getVideo(html)
	if err != nil {
		return nil, err
	}

	video := structure.Video{}
	paths := [][]string{
		{"videoData", "bvid"},
		{"videoData", "aid"},
		{"videoData", "title"},
		{"videoData", "pic"},
		{"videoData", "owner", "mid"},
		{"videoData", "owner", "name"},
		{"videoData", "owner", "face"},
		{"videoData", "desc_v2"},
		{"videoData", "pubdate"},
		{"videoData", "pages"},
		{"videoData", "season_id"},
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
		case 3:
			video.Pic, err = jsonparser.ParseString(value)
			break
		case 4:
			video.Owner.Mid, err = jsonparser.ParseInt(value)
			break
		case 5:
			video.Owner.Name, err = jsonparser.ParseString(value)
			break
		case 6:
			video.Owner.Face, err = jsonparser.ParseString(value)
			break
		case 7:
			_, err = jsonparser.ArrayEach(value, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
				got, err := jsonparser.GetString(value, "raw_text")
				video.Desc = video.Desc + "\n" + got
			})
			break
		case 8:
			video.PubDate, err = jsonparser.ParseInt(value)
			video.PubDate = video.PubDate * 1000
			break
		case 9:
			_, err = jsonparser.ArrayEach(value, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
				videoPage := structure.VideoPage{}
				videoPage.Page, err = jsonparser.GetInt(value, "page")
				videoPage.Part, err = jsonparser.GetString(value, "part")
				videoPage.Duration, err = jsonparser.GetInt(value, "duration")
				video.Pages = append(video.Pages, videoPage)
			})
			break
		case 10:
			video.SeasonId, err = jsonparser.ParseInt(value)
			break
		}
	}, paths...)

	return &video, err
}
