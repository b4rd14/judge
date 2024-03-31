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

func setProblem(problemID string, rds *redis.Client) error {
	ctx := context.Background()
	return rds.Set(ctx, problemID, true, 0).Err()
}

func getProblem(problemID string, rds *redis.Client) (string, error) {
	ctx := context.Background()
	return rds.Get(ctx, problemID).Result()
}
