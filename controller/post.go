package controller

import (
	"islet/dao/redis"
	"islet/logic"
	"islet/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func CreatePostHandler(c *gin.Context) {
	// 1. 获得发帖参数，校验参数
	var post models.Post
	if err := c.ShouldBindJSON(&post); err != nil {
		ResponseErrorWithMsg(c, CodeInvalidParams, err.Error())
		return
	}

	// 2. 从context获取用户id
	userID, err := getCurrentUserID(c)
	if err != nil {
		zap.L().Error("GetCurrentUserID() failed", zap.Error(err))
		ResponseError(c, CodeNotLogin)
		return
	}
	post.AuthorId = userID

	// 3. 调用logic层处理业务
	err = logic.CreatePost(&post)
	if err != nil {
		zap.L().Error("logic.CreatePost() failed", zap.Error(err))
		return
	}

	ResponseSuccess(c, nil)
}

// PostDetailHandler 帖子详情
func PostDetailHandler(c *gin.Context) {
	postID := c.Param("id")

	post, err := logic.GetPost(postID)
	if err != nil {
		zap.L().Error("logic.GetPost() failed", zap.String("postID", postID), zap.Error(err))
		return
	}

	ResponseSuccess(c, post)
}

// GetPostListHandler 查看帖子列表，并根据Redis存储的分数进行排序
// 根据前端传来的参数（分数/创建时间）排序，动态获取帖子列表
// TODO: 按照前端需要的数据，更改返回的posts格式
func GetPostListHandler(c *gin.Context) {
	// 1. 获取请求参数：order 排序方式、page 分页
	//page, size := getPageInfo(c)
	//page, _ := strconv.ParseInt(c.Query("page"), 10, 64)
	order := c.Query("order")
	pageStr, ok := c.GetQuery("page")
	if !ok {
		pageStr = "1"
	}
	page, err := strconv.ParseInt(pageStr, 10, 64)
	if err != nil {
		page = 1
	}

	// 2. 调用logic功能
	posts, err := logic.GetPostList2(order, page)
	if err != nil {
		zap.L().Error("logic.GetPostList2 failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, posts)
}

// PostListHandler 从mysql获取帖子列表（基本功能）
func PostListHandler(c *gin.Context) {
	// 1. 获取分页参数和数据
	page, size := getPageInfo(c)
	data, err := logic.GetPostList(page, size)
	if err != nil {
		zap.L().Error("logic.GetPostList()", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	redis.ShowClient()
	ResponseSuccess(c, data)
}

// VoteHandler 处理帖子投票请求
func VoteHandler(c *gin.Context) {
	// 1. 校验投票信息
	var vote models.VoteData
	if err := c.ShouldBindJSON(&vote); err != nil {
		ResponseErrorWithMsg(c, CodeInvalidParams, err.Error())
		return
	}
	// 获取投票用户id
	userID, err := getCurrentUserID(c)
	if err != nil {
		ResponseError(c, CodeNotLogin)
		return
	}

	// 2.记录投票情况
	if err := logic.PostVote(vote.PostID, userID, vote.Direction); err != nil {
		ResponseError(c, CodeServerBusy)
		return
	}

	ResponseSuccess(c, nil)
}

// getPageInfo 获取分页信息
func getPageInfo(c *gin.Context) (int64, int64) {
	pageStr := c.Query("page")
	sizeStr := c.Query("size")

	var (
		page int64
		size int64
		err  error
	)

	page, err = strconv.ParseInt(pageStr, 10, 64)
	if err != nil {
		page = 1
	}
	size, err = strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		size = 10
	}
	return page, size
}
