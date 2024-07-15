package routes

import (
	"github.com/bilisound/server/internal/api"
	"github.com/bilisound/server/internal/config"
	"github.com/bilisound/server/internal/utils"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func InitRoute(engine *gin.Engine, prefix string) {
	group := engine.Group(prefix)
	{
		group.GET("/metadata", getMetadata)
		group.GET("/resource", getResource)
		group.GET("/test", test)
	}
}

func getMetadata(c *gin.Context) {
	id := c.Query("id")
	video, _, err := api.GetVideoMeta(id, "1")

	if err != nil {
		utils.AjaxError(c, 500, err)
		return
	}
	utils.AjaxSuccess(c, video)
}

func getResource(c *gin.Context) {
	id := c.Query("id")
	episode := c.Query("episode")
	dl := c.Query("dl")
	_, playInfo, err := api.GetVideoMeta(id, episode)

	if err != nil {
		log.Printf("Unable to get resource info: %s\n", err)
		c.Status(500)
		c.Abort()
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", playInfo.Url[0], nil)
	if err != nil {
		log.Printf("Unable to initialize request: %s\n", err)
		c.Status(500)
		c.Abort()
		return
	}

	// Set headers
	rangeValue := c.GetHeader("Range")
	httpCode := 200
	if rangeValue != "" {
		req.Header.Set("Range", c.GetHeader("Range"))
		httpCode = 206
	}
	req.Header.Set("Referer", "https://www.bilibili.com/video/"+id+"/?p="+episode)
	req.Header.Set("User-Agent", config.Global.MustString("request.userAgent"))

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Unable to perform request: %s\n", err)
		c.Status(500)
		c.Abort()
		return
	}
	defer resp.Body.Close()

	// Build Response
	// todo 区分视频和音频，并且正确设置 Content-Disposition 值
	meta := "audio/mp4"
	if dl == "1" {
		c.Header("Content-Disposition", "filename*=utf-8''test.m4a")
		meta = "application/octet-stream"
	}
	c.Header("Accept-Ranges", "bytes")
	c.Header("Cache-Control", "max-age=604800")
	c.Header("Content-Type", meta)
	if rangeValue != "" {
		c.Header("Content-Range", resp.Header.Get("Content-Range"))
		c.Header("Content-Length", resp.Header.Get("Content-Length"))
	}

	c.Status(httpCode)

	// 缓冲区大小
	buffer := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buffer)
		if err != nil {
			break
		}
		if n > 0 {
			if _, err := c.Writer.Write(buffer[:n]); err != nil {
				break
			}
		}
	}
}

func test(c *gin.Context) {
	result, err := api.GetVideoPlayinfo("https://www.bilibili.com/video/BV1ks411D7id/", "2300875", "BV1ks411D7id", "3589854")
	if err != nil {
		utils.AjaxError(c, 500, err)
		return
	}
	utils.AjaxJSONString(c, result)
}
