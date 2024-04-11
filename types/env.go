package types

type ENV struct {
	MinioAccessKey   string `mapstructure:"MINIO_ACCESS_KEY"`
	MinioSecretKey   string `mapstructure:"MINIO_SECRET_KEY"`
	MinioEndpoint    string `mapstructure:"MINIO_ENDPOINT"`
	RabbitmqUsername string `mapstructure:"RABBITMQ_USERNAME"`
	RabbitmqPassword string `mapstructure:"RABBITMQ_PASSWORD"`
	RabbitmqUrl      string `mapstructure:"RABBITMQ_URL"`
}
