package routes

import (
	"github.com/bilisound/server/internal/utils"
	"github.com/gin-gonic/gin"
)

func InitRoute(engine *gin.Engine, prefix string) {
	group := engine.Group(prefix)
	{
		group.GET("/metadata", getMetadata)
	}
}

func getMetadata(c *gin.Context) {
	id := c.Query("id")
	video, err := utils.ParseVideo(id)

	if err != nil {
		utils.AjaxError(c, 500, err)
		return
	}
	utils.AjaxSuccess(c, video)
}
