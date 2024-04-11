package replier

import (
	Type "GO/Judge/types"
	"github.com/spf13/viper"
	"log"
)

type ENV Type.ENV

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
