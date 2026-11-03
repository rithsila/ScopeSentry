package redis

import (
	"context"
	"fmt"
	"os"

	"github.com/Autumn-27/ScopeSentry/internal/logger"

	"github.com/Autumn-27/ScopeSentry/internal/config"
	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func isSkipRedisCommand() bool {
	if len(os.Args) < 2 {
		return false
	}
	switch os.Args[1] {
	case "version", "--version", "-v", "pluginkey", "apikey", "help", "--help", "-h":
		return true
	}
	return false
}

func init() {
	// 创建 Redis 客户端（本地结构体初始化）
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.GlobalConfig.Redis.IP, config.GlobalConfig.Redis.Port),
		Password: config.GlobalConfig.Redis.Password,
		DB:       0,
	})
	Client = client

	if isSkipRedisCommand() {
		return
	}

	// 测试连接
	if err := client.Ping(context.Background()).Err(); err != nil {
		logger.Error(fmt.Sprintf("failed to connect to Redis: %v", err))
		return
	}

	logger.Info("Redis connected successfully")
	return
}

func Close() error {
	if Client != nil {
		if err := Client.Close(); err != nil {
			return fmt.Errorf("failed to close Redis connection: %v", err)
		}
	}
	return nil
}

func SendPluginLogToRedis(key string, msg string) error {
	_, err := Client.RPush(context.Background(), key, msg).Result()
	if err != nil {
		return fmt.Errorf("failed to add values to list: %v", err)
	}
	return nil
}
