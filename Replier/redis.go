package replier

import (
	"context"
	"time"

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

func connectToRedis(ctx context.Context, rds *redis.Client) error {
	return retry(ctx, func() error {
		return rds.Ping(ctx).Err()
	}, 3, time.Second*5)
}

func setResult(ctx context.Context, rds *redis.Client, result []byte, submission SubmissionMessage) error {
	return rds.Set(ctx, "result"+submission.SubmissionID, result, 0).Err()
}

func delResult(ctx context.Context, rds *redis.Client, submission SubmissionMessage) error {
	return rds.Del(ctx, "result"+submission.SubmissionID).Err()
}

func getAllWithPrefix(ctx context.Context, rds *redis.Client, prefix string) ([]string, error) {
	return rds.Keys(ctx, prefix).Result()
}
