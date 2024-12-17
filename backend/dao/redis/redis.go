package redis

import (
	"backend/settings"
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"time"
)

var client *redis.Client

type SliceCmd = redis.SliceCmd
type StringStringMapCmd = redis.StringStringMapCmd

// Init 初始化连接
func Init(cfg *settings.RedisConfig) (err error) {
	client = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password, // no password set
		DB:           cfg.DB,       // use default DB
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
	})

	// 解析默认过期时间
	defaultTTL := time.Duration(cfg.DefaultTTL) * time.Second

	// v8之前的版本：client.Ping().Result()
	_, err = client.Ping(ctx).Result()
	if err != nil {
		return err
	}

	fmt.Printf("Connected to Redis at %s:%d with default TTL: %s\n", cfg.Host, cfg.Port, defaultTTL)
	return nil
}

func Close() {
	_ = client.Close()
}

var ctx = context.Background()

// UpdateTaskStatus 更新任务状态
func UpdateTaskStatus(taskID string, status string, data string) error {
	taskKey := "task_status:" + taskID

	taskData := map[string]interface{}{
		"status":     status,
		"data":       data,
		"updated_at": time.Now().Format(time.RFC3339),
	}
	return client.HMSet(context.Background(), taskKey, taskData).Err()
}

// GetTaskStatus 查询任务状态
func GetTaskStatus(taskID string) (string, map[string]interface{}, error) {
	taskKey := "task_status:" + taskID
	result, err := client.HGetAll(context.Background(), taskKey).Result()
	if err != nil {
		return "", nil, err
	}
	data := map[string]interface{}{}
	if result["data"] != "" {
		data["raw_data"] = result["data"]
	}
	return result["status"], data, nil
}
