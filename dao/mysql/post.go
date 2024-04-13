package mysql

import (
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
	"islet/models"
	"strings"

	"go.uber.org/zap"
)

func CreatePost(post *models.Post) (err error) {
	// 插入帖子信息
	sqlStr := `insert into post(post_id, title, content, author_id, community_id) 
				values(?,?,?,?,?)`
	_, err = db.Exec(sqlStr, post.PostID, post.Title, post.Content, post.AuthorId, post.CommunityID)
	if err != nil {
		zap.L().Error("insert post failed", zap.Error(err))
		err = ErrorInsertFailed
		return
	}
	return
}

func GetPostByID(postID string) (post *models.ApiPostDetail, err error) {
	post = new(models.ApiPostDetail)
	sqlStr := `select post_id, title, content, author_id, community_id, create_time
				from post
				where post_id = ?`
	err = db.Get(post, sqlStr, postID)
	if errors.Is(err, sql.ErrNoRows) {
		err = ErrorInvalidID
		return
	}
	if err != nil {
		zap.L().Error("query post failed", zap.String("postID", postID), zap.Error(err))
		err = ErrorQueryFailed
		return
	}
	return
}

// GetPostByIDs 根据帖子ID列表批量查询 (研究一下sqlx这一块的用法！)
func GetPostByIDs(ids []string) (postList []*models.Post, err error) {
	sqlStr := `select post_id, title, content, author_id, community_id, create_time
				from post
				where post_id in (?)
				order by FIND_IN_SET(post_id, ?)`
	query, args, err := sqlx.In(sqlStr, ids, strings.Join(ids, ","))
	if err != nil {
		return nil, err
	}
	query = db.Rebind(query)
	err = db.Select(&postList, query, args...)
	return
}

func GetPostList(page, size int64) (posts []*models.ApiPostDetail, err error) {
	// 根据page和size取帖子列表信息，并按照创建的时间倒序返回
	sqlStr := `select post_id, title, content, author_id, community_id, create_time
				from post
				order by create_time
				desc
				limit ?,?`
	posts = make([]*models.ApiPostDetail, 0, 2)
	err = db.Select(&posts, sqlStr, (page-1)*size, size)
	return
}
