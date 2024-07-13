package routes

import (
	"github.com/bilisound/server/internal/api"
	"github.com/bilisound/server/internal/utils"
	"github.com/gin-gonic/gin"
)

func InitRoute(engine *gin.Engine, prefix string) {
	group := engine.Group(prefix)
	{
		group.GET("/metadata", getMetadata)
		group.GET("/test", test)
	}
}

func getMetadata(c *gin.Context) {
	id := c.Query("id")
	video, err := api.GetVideoMeta(id)

	if err != nil {
		utils.AjaxError(c, 500, err)
		return
	}
	utils.AjaxSuccess(c, video)
}

func test(c *gin.Context) {
	result, err := api.GetVideoPlayinfo("https://www.bilibili.com/video/BV1ks411D7id/", "2300875", "BV1ks411D7id", "3589854")
	if err != nil {
		utils.AjaxError(c, 500, err)
		return
	}
	utils.AjaxJSONString(c, result)
}
