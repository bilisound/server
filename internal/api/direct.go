package api

import (
	"errors"
	"github.com/bilisound/server/internal/config"
	"github.com/bilisound/server/internal/dao"
	"github.com/bilisound/server/internal/structure"
	"github.com/bilisound/server/internal/utils"
	"github.com/buger/jsonparser"
	"github.com/go-resty/resty/v2"
	"log"
	"regexp"
	"strconv"
)

func parseVideoPlayInfo(playInfo string) (*structure.VideoPlayInfo, error) {
	playInfoByte := []byte(playInfo)
	dashAudio, _, _, dashErr := jsonparser.Get(playInfoByte, "data", "dash", "audio")
	legacyVideo, _, _, legacyErr := jsonparser.Get(playInfoByte, "data", "durl", "[0]")

	if dashErr == nil {
		var maxID int64 = -1
		var maxItemBaseUrl string
		var maxItemBackupUrls []string

		// 遍历 JSON 数组
		_, err := jsonparser.ArrayEach(dashAudio, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			// 获取当前项的 id
			id, err := jsonparser.GetInt(value, "id")
			if err != nil {
				log.Printf("无法获取 id: %v", err)
				return
			}

			// 如果当前项的 id 大于 maxID，则更新 maxID 和相应的 baseUrl 和 backupUrl
			if id > maxID {
				maxID = id
				maxItemBaseUrl, err = jsonparser.GetString(value, "baseUrl")
				if err != nil {
					log.Printf("无法获取 baseUrl: %v", err)
					return
				}

				var backupUrls []string
				_, err = jsonparser.ArrayEach(value, func(urlValue []byte, dataType jsonparser.ValueType, offset int, err error) {
					url, err := jsonparser.ParseString(urlValue)
					if err != nil {
						log.Printf("无法解析 backupUrl: %v", err)
						return
					}
					backupUrls = append(backupUrls, url)
				}, "backupUrl")
				if err != nil {
					log.Printf("无法获取 backupUrl: %v", err)
					return
				}
				maxItemBackupUrls = backupUrls
			}
		})

		if err != nil {
			log.Fatalf("无法解析 JSON 数组: %v", err)
		}

		// 组装新的数组
		result := append([]string{maxItemBaseUrl}, maxItemBackupUrls...)

		return &structure.VideoPlayInfo{
			IsVideo: false,
			Url:     result,
		}, nil
	}

	if legacyErr == nil {
		var baseUrl string
		var backupUrls []string

		baseUrl, err := jsonparser.GetString(legacyVideo, "url")
		if err != nil {
			log.Printf("无法获取 baseUrl: %v", err)
			return nil, err
		}

		_, err = jsonparser.ArrayEach(legacyVideo, func(urlValue []byte, dataType jsonparser.ValueType, offset int, err error) {
			url, err := jsonparser.ParseString(urlValue)
			if err != nil {
				log.Printf("无法解析 backupUrl: %v", err)
				return
			}
			backupUrls = append(backupUrls, url)
		}, "backup_url")
		if err != nil {
			log.Printf("无法获取 backupUrl: %v", err)
			return nil, err
		}

		// 组装新的数组
		result := append([]string{baseUrl}, backupUrls...)

		return &structure.VideoPlayInfo{
			IsVideo: true,
			Url:     result,
		}, nil
	}

	return nil, errors.New("neither dash nor durl can be found")
}

func parseVideoMeta(html string) (video *structure.Video, playInfo *structure.VideoPlayInfo, err error) {
	video = &structure.Video{}
	playInfo = &structure.VideoPlayInfo{}
	initialStateRegex := regexp.MustCompile(`window\.__INITIAL_STATE__=(\{.+);\(function\(\){`)
	playInfoRegex := regexp.MustCompile(`window\.__playinfo__=(\{.+})</script><script>`)

	initialState, err := utils.ExtractContent(initialStateRegex, html, utils.ExtractJSONOptions{})
	if err != nil {
		return nil, nil, err
	}

	initialStateByte := []byte(initialState)

	isFestivalVideo := true
	video.ActivityKey, err = jsonparser.GetString(initialStateByte, "activityKey")
	if err != nil {
		isFestivalVideo = false
		err = nil
	}

	playInfoRaw := ""

	if isFestivalVideo {
		// 节日视频元数据解析
		video.Bvid, err = jsonparser.GetString(initialStateByte, "videoInfo", "bvid")
		_, err := jsonparser.ArrayEach(initialStateByte, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			bvid, err := jsonparser.GetString(value, "bvid")
			if bvid == video.Bvid {
				video.Pic, err = jsonparser.GetString(value, "cover")
				video.Owner.Name, err = jsonparser.GetString(value, "author", "name")
				video.Owner.Face, err = jsonparser.GetString(value, "author", "face")
				video.Owner.Mid, err = jsonparser.GetInt(value, "author", "mid")
			}
		}, "sectionEpisodes")

		if err != nil {
			return nil, nil, err
		}

		paths := [][]string{
			{"videoInfo", "bvid"},
			{"videoInfo", "aid"},
			{"videoInfo", "title"},
			{"videoInfo", "desc"},
			{"videoInfo", "pubdate"},
			{"videoInfo", "pages"},
		}

		jsonparser.EachKey(initialStateByte, func(idx int, value []byte, vt jsonparser.ValueType, err error) {
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
					videoPage.Cid, err = jsonparser.GetInt(value, "cid")
					video.Pages = append(video.Pages, videoPage)
				})
				break
			}
		}, paths...)
		if err != nil {
			return nil, nil, err
		}

		// 节日视频播放信息提取
		playInfoRaw, err = GetVideoPlayinfo("", strconv.FormatInt(video.Aid, 10), video.Bvid, strconv.FormatInt(video.Pages[0].Cid, 10))
		if err != nil {
			return nil, nil, err
		}
	} else {
		// 常规视频元数据解析
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

		jsonparser.EachKey(initialStateByte, func(idx int, value []byte, vt jsonparser.ValueType, err error) {
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
					videoPage.Cid, err = jsonparser.GetInt(value, "cid")
					video.Pages = append(video.Pages, videoPage)
				})
				break
			case 10:
				video.SeasonId, err = jsonparser.ParseInt(value)
				break
			}
		}, paths...)
		if err != nil {
			return nil, nil, err
		}

		// 常规视频播放信息提取
		playInfoRaw, err = utils.ExtractContent(playInfoRegex, html, utils.ExtractJSONOptions{})
		if err != nil {
			return nil, nil, err
		}
	}

	parsed, err := parseVideoPlayInfo(playInfoRaw)
	if parsed.IsVideo {
		video.IsVideoOnly = true
	}
	return video, parsed, err
}

func GetVideoMeta(id string, episode string) (video *structure.Video, playinfo *structure.VideoPlayInfo, err error) {
	got, err := dao.GetCache("GetVideoMeta_" + id)
	if err != nil {
		return nil, nil, err
	}

	// 没缓存
	if got == "" {
		log.Println("Cache not found, requesting from external API")
		// Create a Resty Client
		client := resty.New()

		resp, err := client.R().
			EnableTrace().
			SetHeader("User-Agent", config.Global.MustString("request.userAgent")).
			Get("https://www.bilibili.com/video/" + id + "/?p=" + episode)

		if err != nil {
			return nil, nil, err
		}

		got = resp.String()

		err = dao.SetCache("GetVideoMeta_"+id, got)
		if err != nil {
			return nil, nil, err
		}
	}

	video, playInfo, err := parseVideoMeta(got)
	return video, playInfo, err
}
