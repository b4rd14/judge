package replier

import (
	"context"

	redis "github.com/redis/go-redis/v9"
)

func NewRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0})
}

func setProblem(ctx context.Context, rds *redis.Client, problemID string) error {
	return rds.Set(ctx, problemID, true, 0).Err()
}

func getProblem(ctx context.Context, rds *redis.Client, problemID string) (string, error) {
	return rds.Get(ctx, problemID).Result()
}
