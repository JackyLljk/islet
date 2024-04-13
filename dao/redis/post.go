package redis

import (
	"errors"
	"github.com/redis/go-redis/v9"
	"islet/models"
	"math"
	"strconv"
	"time"
)

var (
	ErrorVoteTimeExpire = errors.New("已过投票时间")
	ErrorVoted          = errors.New("已经投过票了")
)

// 存储Redis用到的key
// 用":"分割为不同的命名空间
// Redis key注意使用命名空间的方式，方便查询和拆分
// 可以将类型写在常量名上
const (
	KeyPostTimeZSet  = "islet:post:time"  // 以时间为分数的Zset1
	KeyPostScoreZSet = "islet:post:score" // 以投票数为分数的Zset（投票数*432）

	// KeyPostVotedZSetPrefix 记录给帖子投票结果的前缀：投票的用户，投票的分数 (UserID, score)
	// Prefix 结尾，说明后续操作时会补充key
	KeyPostVotedZSetPrefix = "islet:post:voted:"

	// KeyPostInfoHashPrefix 记录帖子信息的哈希表Hash
	KeyPostInfoHashPrefix = "islet:post:"

	// KeyCommunityPostSetPrefix 记录社区帖子的集合Set
	KeyCommunityPostSetPrefix = "islet:community:"

	// KeyPostVotedZSetPrefix KeyPostVotedUpSetPrefix   = "bluebell:post:voted:down:"
	// KeyPostVotedDownSetPrefix = "bluebell:post:voted:up:"
)

const (
	OneWeekInSeconds = 7 * 24 * 3600
	// VoteScore 每一票对应的分数
	VoteScore  float64 = 432
	PostPerAge         = 20
)

// CreatePost 使用Hash存储帖子信息
func CreatePost(postID, communityID uint64) (err error) {
	// 事务操作（确保操作同时成功）
	pipeline := client.TxPipeline()

	pipeline.ZAdd(ctx, KeyPostTimeZSet, redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: postID,
	})

	pipeline.ZAdd(ctx, KeyPostScoreZSet, redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: postID,
	})

	cKey := KeyCommunityPostSetPrefix + strconv.Itoa(int(communityID))
	pipeline.SAdd(ctx, cKey, postID)
	_, err = pipeline.Exec(ctx)

	return
}

// PostVote 为帖子投票，v = -1/0/1
func PostVote(postID, userID string, v float64) (err error) {
	// 1. 取帖子发布时间 islet:post:time
	// 取集合 islet:post:time 中 postID 的分数值，即时间
	postTime := client.ZScore(ctx, KeyPostTimeZSet, postID).Val()
	// 帖子发布时间超过一周，已过投票时间
	if float64(time.Now().Unix())-postTime > OneWeekInSeconds {
		// 不允许投票了
		return ErrorVoteTimeExpire
	}

	// 判断是否已经投过票
	key := KeyPostVotedZSetPrefix + postID
	// 查用户给当前帖子的投票记录
	ov := client.ZScore(ctx, key, userID).Val()

	// 计算当前投票和之前投票的差值
	diffAbs := math.Abs(ov - v) // 计算两次投票的差值

	// 开启事务，整合多个命令发送到 Redis 服务端
	// TxPipeline 可以理解为支持事务的 Pipeline，简单理解为可以保证事务的原子性
	pipeline := client.TxPipeline()

	// 记录本次投票到有序集合 islet:post:voted:postID
	pipeline.ZAdd(ctx, key, redis.Z{ // 记录已投票
		Score:  v,
		Member: userID,
	})

	/* 投票后更新帖子分数，分数修改规则如下

	v=0：取消投票，分数不变

	diffAbs=0：投票没有更改，分数不变

	v=1：投赞成票
		1. 之前投反对票，现在投赞成票：diffAbs=2，分数增加 432*2*1 = 864
		2. 之前没投票，现在投赞成票：diffAbs=1，分数增加 432*1 = 432

	v=-1：投反对票
		1. 之前投赞成票，现在投反对票：diffAbs=2，分数减少 432*2*(-1) = -864
		2. 之前没投票，现在投反对票：diffAbs=1，分数减少 432*(-1) = -432

	*/
	pipeline.ZIncrBy(ctx, KeyPostScoreZSet, VoteScore*diffAbs*v, postID)

	// 投票后更新帖子得票信息：以用户给帖子前后投票情况，增减帖子的总票数(只记数量，不计正反)
	switch math.Abs(ov) - math.Abs(v) {
	case 1:
		// 取消投票 ov=1/-1 v=0
		// 投票数-1
		pipeline.HIncrBy(ctx, KeyPostInfoHashPrefix+postID, "votes", -1)
	case 0:
		// 反转投票 ov=-1/1 v=1/-1，票数还是没有变化，不做更改
	case -1:
		// 新增投票 ov=0 v=1/-1
		// 投票数+1
		pipeline.HIncrBy(ctx, KeyPostInfoHashPrefix+postID, "votes", 1)
	default:
		// 已经投过票了
		return ErrorVoted
	}

	// 结束事务
	_, err = pipeline.Exec(ctx)
	return
}

// GetPostIdsInOrder 从Redis中获取查询范围内的postID
func GetPostIdsInOrder(order string, page int64) ([]string, error) {
	key := KeyPostScoreZSet
	if order == models.OrderTime {
		key = KeyPostTimeZSet
	}
	start := (page - 1) * PostPerAge
	end := start + PostPerAge - 1
	return client.ZRevRange(ctx, key, start, end).Result()
}

// GetPostsAgreeVoteData 根据帖子列表查询每篇帖子的投赞成票的数据
func GetPostsAgreeVoteData(ids []string) (data []int64, err error) {
	pipeline := client.Pipeline()
	for _, id := range ids {
		key := KeyPostVotedZSetPrefix + id
		pipeline.ZCount(ctx, key, "1", "1")
	}
	cmders, err := pipeline.Exec(ctx)
	if err != nil {
		return nil, err
	}
	data = make([]int64, 0, len(cmders))
	for _, cmder := range cmders {
		v := cmder.(*redis.IntCmd).Val()
		data = append(data, v)
	}
	return
}

//func GetCommunityPostIdsInOrder(p *models.ParamPostList) ([]string, error) {
//	key := KeyPostTimeZSet
//	if p.Order == models.OrderScore {
//		key = KeyPostScoreZSet
//	}
//
//	// 1. 按照社区分类，创造一个新的zset，存储社区内的帖子及其分数
//	cKey := KeyCommunityPostSetPrefix + strconv.Itoa(int(p.CommunityID))
//
//	// 利用缓存k减少社区集合执行的次数
//	k := key + strconv.Itoa(int(p.CommunityID))
//	if client.Exists(ctx, key).Val() < 1 {
//		// 不存在，需要计算
//		pipeline := client.Pipeline()
//		pipeline.ZInterStore(ctx, key, &redis.ZStore{
//			Aggregate: "MAX",
//		})
//		pipeline.Expire(ctx, key, 60*time.Second) // 设置超时时间
//		_, err := pipeline.Exec(ctx)
//		if err != nil {
//			return nil, err
//		}
//	}
//	// 存在的话就直接根据key查询ids
//	start := (p.Page - 1) * p.Size
//	end := start + p.Size - 1
//	return client.ZRevRange(ctx, k, start, end).Result()
//}

// GetPost 从key中分页取出帖子
//func GetPost(order string, page int64) []map[string]string {
//	// 1. 根据order选择从不同的zset进行查询(默认按照分数排序)
//	key := KeyPostScoreZSet
//	if order == "time" {
//		key = KeyPostTimeZSet
//	}
//
//	// 2. 返回查询页范围内的记录，拿到的是postID和score
//	start := (page - 1) * PostPerAge
//	end := start + PostPerAge - 1
//	ids := client.ZRevRange(ctx, key, start, end).Val()
//
//	// 3. 根据postID查询哈希表，得到存储的帖子详细信息
//	postList := make([]map[string]string, 0, len(ids))
//	for _, id := range ids {
//		postData := client.HGetAll(ctx, KeyPostInfoHashPrefix+id).Val()
//		postData["id"] = id
//		postList = append(postList, postData)
//	}
//	return postList
//}

// GetCommunityPost 分社区根据发帖时间或者分数取出分页的帖子
//func GetCommunityPost(communityName, orderKey string, page int64) []map[string]string {
//	key := orderKey + communityName // 创建缓存键
//
//	if client.Exists(ctx, key).Val() < 1 {
//		client.ZInterStore(ctx, key, &redis.ZStore{
//			Aggregate: "MAX",
//		}, KeyCommunityPostSetPrefix+communityName, orderKey)
//		client.Expire(ctx, key, 60*time.Second)
//	}
//	return GetPost(key, page)
//}
