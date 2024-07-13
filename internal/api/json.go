package api

import (
	"github.com/bilisound/server/internal/config"
	"github.com/bilisound/server/internal/utils"
	"github.com/go-resty/resty/v2"
	"net/url"
)

var jsonClient = resty.New()

func init() {
	jsonClient.OnBeforeRequest(func(c *resty.Client, req *resty.Request) error {
		req.SetHeader("User-Agent", config.Global.MustString("request.userAgent"))
		var err error
		req.URL, err = utils.SignAndGenerateURL(req.URL, req.QueryParam)
		if err != nil {
			return err
		}
		req.QueryParam = url.Values{}
		req.URL = "https://api.bilibili.com" + req.URL
		return nil
	})
}

func GetVideoPlayinfo(referer string, avid string, bvid string, cid string) (string, error) {
	resp, err := jsonClient.R().
		EnableTrace().
		SetQueryParam("avid", avid).
		SetQueryParam("bvid", bvid).
		SetQueryParam("cid", cid).
		SetHeader("referer", referer).
		Get("/x/player/wbi/playurl")

	return resp.String(), err
}
