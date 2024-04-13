package logic

import (
	"fmt"
	"go.uber.org/zap"
	"islet/dao/mysql"
	"islet/dao/redis"
	"islet/models"
	"islet/pkg/snowflake"
)

func CreatePost(post *models.Post) (err error) {
	// 1. 雪花算法生成帖子id
	postID, err := snowflake.GetId()
	if err != nil {
		zap.L().Error("snowflake.GetId() failed", zap.Error(err))
		return
	}
	post.PostID = postID

	// 2. 创建帖子
	if err := mysql.CreatePost(post); err != nil {
		zap.L().Error("mysql.CreatePost() failed", zap.Error(err))
		return err
	}

	// 3. 在Redis中创建帖子
	if err := redis.CreatePost(post.PostID, post.CommunityID); err != nil {
		zap.L().Error("redis.CreatePost() failed", zap.Error(err))
		return err
	}
	return
}

// GetPost 根据帖子id获取帖子详情，同时补齐帖子的一些信息
func GetPost(postID string) (post *models.ApiPostDetail, err error) {
	// 1. 通过post_id查找帖子详情
	post, err = mysql.GetPostByID(postID)
	if err != nil {
		zap.L().Error("mysql.GetPostByID() failed", zap.String("post_id", postID), zap.Error(err))
		return
	}

	// 2. 查找发帖人（用户名）
	// 将id转为string
	user, err := mysql.GetUserByID(fmt.Sprint(post.AuthorId))
	if err != nil {
		zap.L().Error("mysql.GetUserByID() failed", zap.String("author_id", fmt.Sprint(post.AuthorId)), zap.Error(err))
		return
	}
	post.AuthorName = user.Username

	// 3. 查找社区名
	community, err := mysql.GetCommunityDetail(fmt.Sprint(post.CommunityID))
	if err != nil {
		zap.L().Error("mysql.GetCommunityByID() failed", zap.String("community_id", fmt.Sprint(post.CommunityID)), zap.Error(err))
		return
	}
	post.CommunityName = community.CommunityName

	return post, nil
}

func GetPostList(page, size int64) (data []*models.ApiPostDetail, err error) {
	// 1. 查询帖子表记录
	postList, err := mysql.GetPostList(page, size)
	if err != nil {
		fmt.Println(err)
		return
	}

	data = make([]*models.ApiPostDetail, 0, len(postList))
	// 2. 补齐帖子详情api ApiPostDetail
	for _, post := range postList {
		user, err := mysql.GetUserByID(fmt.Sprint(post.AuthorId))
		if err != nil {
			zap.L().Error("mysql.GetUserByID() failed", zap.String("author_id", fmt.Sprint(post.AuthorId)), zap.Error(err))
			continue
		}
		post.AuthorName = user.Username

		community, err := mysql.GetCommunityDetail(fmt.Sprint(post.CommunityID))
		if err != nil {
			zap.L().Error("mysql.GetCommunityByID() failed", zap.String("community_id", fmt.Sprint(post.CommunityID)), zap.Error(err))
			continue
		}
		post.CommunityName = community.CommunityName
		data = append(data, post)
	}
	return
}

// GetPostListRedis 根据社区调用查询帖子列表的接口
//func GetPostListRedis(p *models.ParamPostList) (data []*models.ApiPostDetail2, err error) {
//	if p.CommunityID == 0 {
//		data, err = GetPostList2(p)
//	} else {
//		data, err = GetCommunityPostList(p)
//	}
//	return
//}

// GetPostList2 在所有帖子中查询
func GetPostList2(order string, page int64) (data []*models.ApiPostDetail2, err error) {
	// 1. 根据order和page得到查询范围内的postID
	ids, err := redis.GetPostIdsInOrder(order, page)
	if err != nil {
		return
	}
	// 没有查到数据
	if len(ids) == 0 {
		zap.L().Warn("redis.GetPostIdsInOrder() return 0 data")
		return
	}

	// 2. 根据postID从MySQL中取得帖子详细信息
	posts, err := mysql.GetPostByIDs(ids)
	if err != nil {
		return
	}
	zap.L().Debug("GetPostList2", zap.Any("posts", posts))
	voteData, err := redis.GetPostsAgreeVoteData(ids)
	if err != nil {
		return
	}

	// 3. 将帖子的作者及分区信息查询出来填充到帖子中
	for idx, post := range posts {
		user, err := mysql.GetUserByID(fmt.Sprint(post.AuthorId))
		if err != nil {
			zap.L().Error("mysql.GetUserByID failed",
				zap.Uint64("community_id", post.CommunityID),
				zap.Error(err))
			continue
		}
		community, err := mysql.GetCommunityDetail(fmt.Sprint(post.CommunityID))
		if err != nil {
			zap.L().Error("mysql.GetCommunityDetail failed",
				zap.Uint64("community_id", post.CommunityID),
				zap.Error(err))
			continue
		}
		postDetail := &models.ApiPostDetail2{
			AuthorName:      user.Username,
			VoteNum:         voteData[idx],
			Post:            post,
			CommunityDetail: community,
		}
		data = append(data, postDetail)
	}
	return data, err
}

// GetCommunityPostList 在对应社区中进行查询
//func GetCommunityPostList(p *models.ParamPostList) (data []*models.ApiPostDetail2, err error) {
//	// 1. 根据order和page得到查询范围内的postID
//	ids, err := redis.GetCommunityPostIdsInOrder(p)
//	if err != nil {
//		return
//	}
//	// 没有查到数据
//	if len(ids) == 0 {
//		zap.L().Warn("redis.GetCommunityPostIdsInOrder() return 0 data")
//		return
//	}
//
//	// 2. 根据postID从MySQL中取得帖子详细信息
//	posts, err := mysql.GetPostByIDs(ids)
//	if err != nil {
//		return
//	}
//	zap.L().Debug("GetPostList2", zap.Any("posts", posts))
//	voteData, err := redis.GetPostsAgreeVoteData(ids)
//	if err != nil {
//		return
//	}
//
//	// 3. 将帖子的作者及分区信息查询出来填充到帖子中
//	for idx, post := range posts {
//		user, err := mysql.GetUserByID(fmt.Sprint(post.AuthorId))
//		if err != nil {
//			zap.L().Error("mysql.GetUserByID failed",
//				zap.Uint64("community_id", post.CommunityID),
//				zap.Error(err))
//			continue
//		}
//		community, err := mysql.GetCommunityDetail(fmt.Sprint(post.CommunityID))
//		if err != nil {
//			zap.L().Error("mysql.GetCommunityDetail failed",
//				zap.Uint64("community_id", post.CommunityID),
//				zap.Error(err))
//			continue
//		}
//		postDetail := &models.ApiPostDetail2{
//			AuthorName:      user.Username,
//			VoteNum:         voteData[idx],
//			Post:            post,
//			CommunityDetail: community,
//		}
//		data = append(data, postDetail)
//	}
//	return
//}

// PostVote 处理投票业务
func PostVote(postID string, userID uint64, v float64) (err error) {
	return redis.PostVote(postID, fmt.Sprint(userID), v)
}
