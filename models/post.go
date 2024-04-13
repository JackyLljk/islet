package models

import "time"

const (
	OrderTime  = "time"
	OrderScore = "score"
)

type Post struct {
	PostID      uint64    `json:"post_id" db:"post_id"`
	Title       string    `json:"title" db:"title"`
	Content     string    `json:"content" db:"content"`
	AuthorId    uint64    `json:"author_id" db:"author_id"`
	CommunityID uint64    `json:"community_id" db:"community_id"`
	Status      int32     `json:"status" db:"status"`
	CreateTime  time.Time `json:"-" db:"create_time"`
}

// VoteData 帖子投票信息
type VoteData struct {
	// PostID    string  `json:"post_id,string"`，注意可能的id参数失真问题
	PostID    string  `json:"post_id"`
	Direction float64 `json:"direction"`
}

// ApiPostDetail 帖子详情接口，获得帖子信息，以及发帖人和社区名等额外信息
type ApiPostDetail struct {
	*Post
	AuthorName    string `json:"author_name"`
	CommunityName string `json:"community_name"`
}

type ApiPostDetail2 struct {
	AuthorName       string             `json:"author_name"`
	VoteNum          int64              `json:"vote_num"`
	*Post                               // 嵌入帖子结构体
	*CommunityDetail `json:"community"` // 嵌入社区信息
}

type ParamPostList struct {
	CommunityID int64  `json:"community_id" form:"community_id"` // 可以为空
	Page        int64  `json:"page" form:"page"`
	Size        int64  `json:"size" form:"size"`
	Order       string `json:"order" form:"order"`
}
