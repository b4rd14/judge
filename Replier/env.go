package replier

import (
	"github.com/spf13/viper"
	"log"
)

type ENV struct {
	MinioAccessKey   string `mapstructure:"MINIO_ACCESS_KEY"`
	MinioSecretKey   string `mapstructure:"MINIO_SECRET_KEY"`
	MinioEndpoint    string `mapstructure:"MINIO_ENDPOINT"`
	RabbitmqUsername string `mapstructure:"RABBITMQ_USERNAME"`
	RabbitmqPassword string `mapstructure:"RABBITMQ_PASSWORD"`
	RabbitmqUrl      string `mapstructure:"RABBITMQ_URL"`
}

func NewEnv() *ENV {
	env := ENV{}
	viper.SetConfigFile(".env")
	err := viper.ReadInConfig()
	if err != nil {
		log.Println("Can't find the file .env : ", err)
	}
	err = viper.Unmarshal(&env)
	if err != nil {
		log.Println("Environment can't be loaded: ", err)
	}
	return &env
}
