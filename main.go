package main

import (
	"fmt"
	"islet/dao/mysql"
	"islet/dao/redis"
	"islet/logger"
	"islet/pkg/snowflake"
	"islet/routers"
	"islet/settings"
)

// 程序的入口，较通用的go web开发脚手架模板
// TODO: 尝试对所有模块的函数接口编写单元测试用例，创建一个测试分支（练手测试）

func main() {
	// init settings
	if err := settings.Init(); err != nil {
		fmt.Printf("init settings failed, err:%v\n", err)
		return
	}

	// init log
	if err := logger.Init(settings.Conf.LogConfig, settings.Conf.Mode); err != nil {
		fmt.Printf("init logger failed, err:%v\n", err)
		return
	}

	// init mysql
	if err := mysql.Init(settings.Conf.MySQLConfig); err != nil {
		fmt.Printf("init mysql failed, err:%v\n", err)
		return
	}
	defer mysql.Close()

	if err := snowflake.Init(settings.Conf.MachineID, settings.Conf.StartTime); err != nil {
		fmt.Printf("init snowflake failed, err:%v\n", err)
		return
	}

	// init redis
	if err := redis.Init(settings.Conf.RedisConfig); err != nil {
		fmt.Printf("init redis failed, err:%v\n", err)
		return
	}
	defer redis.Close()

	//  register router
	r := routers.SetupRouter()
	err := r.Run(fmt.Sprintf(":%d", settings.Conf.Port))
	if err != nil {
		fmt.Printf("run server failed, err: %v\n", err)
		return
	}
}
