package api

import (
	"github.com/bilisound/server/internal/config"
	"github.com/bilisound/server/internal/structure"
	"github.com/bilisound/server/internal/utils"
	"github.com/buger/jsonparser"
	"github.com/go-resty/resty/v2"
	"regexp"
)

func parseRegularVideoMeta(html string) (*structure.Video, error) {
	initialStateRegex := regexp.MustCompile(`window\.__INITIAL_STATE__=(\{.+);\(function\(\){`)
	//playInfoRegex := regexp.MustCompile(`window\.__playinfo__=(\{.+})</script><script>`)

	initialState, err := utils.ExtractContent(initialStateRegex, html, utils.ExtractJSONOptions{})
	if err != nil {
		return nil, err
	}

	isFestivalVideo := true
	_, err = jsonparser.GetString([]byte(initialState), "activityKey")
	if err != nil {
		isFestivalVideo = false
		err = nil
	}

	//playInfo, err := utils.ExtractContent(playInfoRegex, html, utils.ExtractJSONOptions{})
	//if err != nil {
	//	return nil, err
	//}

	video := structure.Video{}
	video.Bvid, err = jsonparser.GetString([]byte(initialState), "videoInfo", "bvid")

	if isFestivalVideo {
		_, err := jsonparser.ArrayEach([]byte(initialState), func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			bvid, err := jsonparser.GetString(value, "bvid")
			if bvid == video.Bvid {
				video.Pic, err = jsonparser.GetString(value, "cover")
				video.Owner.Name, err = jsonparser.GetString(value, "author", "name")
				video.Owner.Face, err = jsonparser.GetString(value, "author", "face")
				video.Owner.Mid, err = jsonparser.GetInt(value, "author", "mid")
			}
		}, "sectionEpisodes")

		if err != nil {
			return nil, err
		}

		paths := [][]string{
			{"videoInfo", "bvid"},
			{"videoInfo", "aid"},
			{"videoInfo", "title"},
			{"videoInfo", "desc"},
			{"videoInfo", "pubdate"},
			{"videoInfo", "pages"},
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
				video.Desc, err = jsonparser.ParseString(value)
				break
			case 4:
				video.PubDate, err = jsonparser.ParseInt(value)
				video.PubDate = video.PubDate * 1000
				break
			case 5:
				_, err = jsonparser.ArrayEach(value, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
					videoPage := structure.VideoPage{}
					videoPage.Page, err = jsonparser.GetInt(value, "page")
					videoPage.Part, err = jsonparser.GetString(value, "part")
					videoPage.Duration, err = jsonparser.GetInt(value, "duration")
					video.Pages = append(video.Pages, videoPage)
				})
				break
			}
		}, paths...)
		if err != nil {
			return nil, err
		}
	} else {
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
		if err != nil {
			return nil, err
		}
	}

	return &video, err
}

func GetVideoMeta(id string) (*structure.Video, error) {
	// Create a Resty Client
	client := resty.New()

	resp, err := client.R().
		EnableTrace().
		SetHeader("User-Agent", config.Global.MustString("request.userAgent")).
		Get("https://www.bilibili.com/video/" + id + "/")

	if err != nil {
		return nil, err
	}

	result, err := parseRegularVideoMeta(resp.String())
	return result, err
}
