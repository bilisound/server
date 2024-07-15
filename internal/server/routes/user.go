package routes

import (
	"fmt"
	"github.com/bilisound/server/internal/api"
	"github.com/bilisound/server/internal/config"
	"github.com/bilisound/server/internal/utils"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

func InitRoute(engine *gin.Engine, prefix string) {
	group := engine.Group(prefix)
	{
		group.GET("/metadata", getMetadata)
		group.GET("/resource", getResource)
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
	initialState, playInfo, err := api.GetVideoMeta(id, episode)

	if err != nil {
		log.Printf("Unable to get resource info: %s\n", err)
		c.Status(500)
		c.Abort()
		return
	}

	// 构建请求
	log.Printf("Retriving resources from %s\n", playInfo.Url[0])
	client := &http.Client{}
	req, err := http.NewRequest("GET", playInfo.Url[0], nil)
	if err != nil {
		log.Printf("Unable to initialize request: %s\n", err)
		c.Status(500)
		c.Abort()
		return
	}

	// 设置请求头部
	rangeValue := c.GetHeader("Range")
	httpCode := 200
	if rangeValue != "" {
		req.Header.Set("Range", c.GetHeader("Range"))
		httpCode = 206
	}
	req.Header.Set("Referer", "https://www.bilibili.com/video/"+id+"/?p="+episode)
	req.Header.Set("User-Agent", config.Global.MustString("request.userAgent"))

	// 发出请求
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Unable to perform request: %s\n", err)
		c.Status(500)
		c.Abort()
		return
	}
	defer resp.Body.Close()

	// 构建响应
	if dl == "" {
		if initialState.IsVideoOnly {
			c.Header("Content-Type", "video/mp4")
		} else {
			c.Header("Content-Type", "audio/mp4")
		}
	} else {
		// 构建文件名
		fileName := "["
		if dl == "av" {
			fileName += "av" + strconv.FormatInt(initialState.Aid, 10)
		} else {
			fileName += initialState.Bvid
		}
		episodeNum, err := strconv.Atoi(episode)
		if err != nil {
			log.Printf("Unable to construct file name: %s\n", err)
			c.Status(500)
			c.Abort()
			return
		}
		fileName += "] [P" + episode + "] " + initialState.Pages[episodeNum-1].Part
		if initialState.IsVideoOnly {
			fileName += ".mp4"
		} else {
			fileName += ".m4a"
		}

		// 使用 RFC 5987 编码文件名
		encodedFileName := url.PathEscape(fileName)
		contentDisposition := fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s", fileName, encodedFileName)
		c.Header("Content-Disposition", contentDisposition)
		c.Header("Content-Type", "application/octet-stream")
	}
	c.Header("Accept-Ranges", "bytes")
	c.Header("Cache-Control", "max-age=604800")
	if rangeValue != "" {
		c.Header("Content-Range", resp.Header.Get("Content-Range"))
		c.Header("Content-Length", resp.Header.Get("Content-Length"))
	}

	c.Status(httpCode)

	// 数据传输
	_, err = io.Copy(c.Writer, resp.Body)
	if err != nil {
		log.Printf("Unable to transfer data: %s\n", err)
		c.Abort()
		return
	}
	log.Printf("Success transfer data")
}
