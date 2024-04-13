package controller

import (
	"islet/logic"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CommunityHandler 社区列表
func CommunityHandler(c *gin.Context) {
	communityList, err := logic.GetCommunityList()

	if err != nil {
		zap.L().Error("mysql.GetCommunityList() failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}

	ResponseSuccess(c, communityList)
}

// CommunityDetailHandler 社区详情
func CommunityDetailHandler(c *gin.Context) {
	communityID := c.Param("id")
	communityList, err := logic.GetCommunityDetail(communityID)
	if err != nil {
		zap.L().Error("mysql.GetCommunityDetail() failed", zap.Error(err))
		ResponseErrorWithMsg(c, CodeSuccess, err.Error())
		return
	}
	ResponseSuccess(c, communityList)
}
