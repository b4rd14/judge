package replier

import (
	model "GO/Judge/Model"
	"github.com/spf13/viper"
	"log"
)

func NewEnv() *model.ENV {
	env := model.ENV{}
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
