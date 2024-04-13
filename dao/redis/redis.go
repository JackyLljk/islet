package redis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"islet/settings"
)

var (
	client *redis.Client
	Nil    = redis.Nil
	ctx    = context.Background()
)

func ShowClient() {
	fmt.Println(client)
}

//type SliceCmd = redis.SliceCmd
//type StringStringMapCmd = redis.StringStringMapCmd

// Init 初始化Redis连接
func Init(cfg *settings.RedisConfig) (err error) {
	host := cfg.Host
	password := cfg.Password
	database := cfg.DB
	port := cfg.Port

	// 连接Redis
	client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: password,
		DB:       database,
	})

	_, err = client.Ping(ctx).Result()
	if err != nil {
		return err
	}
	zap.L().Info("redis init success")
	fmt.Println(client)
	return nil
}

func Close() {
	_ = client.Close()
}
