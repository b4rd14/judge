package Model

import "time"

type ENV struct {
	MinioAccessKey   string `mapstructure:"MINIO_ACCESS_KEY"`
	MinioSecretKey   string `mapstructure:"MINIO_SECRET_KEY"`
	MinioEndpoint    string `mapstructure:"MINIO_ENDPOINT"`
	RabbitmqUsername string `mapstructure:"RABBITMQ_USERNAME"`
	RabbitmqPassword string `mapstructure:"RABBITMQ_PASSWORD"`
	RabbitmqUrl      string `mapstructure:"RABBITMQ_URL"`
}

type SubmissionMessage struct {
	SubmissionID   string
	ProblemID      string
	UserID         string
	TimeStamp      string
	Type           string
	TestCaseNumber int
	TimeLimit      time.Duration
	MemoryLimit    int64
}

type SubmissionResult struct {
	SubmissionID string
	ProblemID    string
	Output       string
}

type MinioConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	UseSSL    bool
}
