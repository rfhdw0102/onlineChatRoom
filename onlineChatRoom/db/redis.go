package db

import (
	"fmt"
	"github.com/go-redis/redis"
	"strings"
)

var RDB *redis.Client

// InitRedis 连接Redis
func InitRedis() error {
	RDB = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		PoolSize: 100, //连接池的大小
	})
	_, err := RDB.Ping().Result()
	if err != nil {
		return fmt.Errorf("rdb.Ping() failed:%w", err)
	}
	return nil
}

// AddActivity 给用户追加活跃度
func AddActivity(username string, number float64) error {
	err := RDB.ZIncrBy("activityRank", number, username).Err()
	if err != nil {
		return fmt.Errorf("rdb.ZIncrBy failed:%w", err)
	}
	return nil
}

// ShowActivityRank 显示活跃度排名
func ShowActivityRank() (string, error) {
	// zSlice 是一个结构体，存放排名信息
	zSlice, err := RDB.ZRevRangeWithScores("activityRank", 0, -1).Result()
	if err != nil {
		return "", fmt.Errorf("rdb.ZRevRangeWithScores failed:%w", err)
	}
	var sprintf string
	for i, value := range zSlice {
		if value.Member.(string) == "系统广播" {
			continue
		}
		// 显示排名、名字和分数、排名从 1 开始
		sprintf = fmt.Sprintf("%s排名 %d: %s\t, 活跃度=%d\n", sprintf, i+1, value.Member.(string), int(value.Score))
	}
	return strings.Trim(sprintf, "\n"), nil
}
