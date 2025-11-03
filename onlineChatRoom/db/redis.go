package db

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"log"
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

// AddStreamsData 向streams流中添加数据
func AddStreamsData(username string, content string, receiver string) (string, error) {
	msgID, err := RDB.XAdd(&redis.XAddArgs{
		Stream: "room", // 接收都用这一个streams流
		MaxLen: 100,    // 限制最大消息长度，超出自动清除
		Values: map[string]interface{}{
			"sender":   username,
			"content":  content,
			"receiver": receiver,
		},
	}).Result()
	if err != nil {
		return "", fmt.Errorf("streams添加数据失败:%w", err)
	}
	return msgID, nil
}

// ReadStreams 读取流消息
func ReadStreams(count int64, method string) ([]redis.XMessage, error) {
	result, err := RDB.XRead(&redis.XReadArgs{
		Streams: []string{"room", method},
		Count:   count,
		Block:   0,
	}).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("XREAD error: %w", err)
	}
	return result[0].Messages, nil
}

// ShowHistory 查看历史消息,限制只能查看10条历史消息
func ShowHistory() (string, error) {
	res, err := RDB.XRevRangeN("room", "+", "-", 10).Result()
	if err != nil {
		return "", fmt.Errorf("XRevRangeN failed:%w", err)
	}
	var history string
	for i := len(res) - 1; i >= 0; i-- {
		values := res[i].Values
		history += fmt.Sprintf("%s: %s\n", values["sender"], values["content"])
	}
	return history, nil
}

// ClearRedis 服务端重启时清空活跃度排行和streams流
func ClearRedis() {
	err := RDB.Del("room", "activityRank").Err()
	if err != nil {
		log.Println("重新开启服务端时清空Redis数据失败:", err)
	}
}
