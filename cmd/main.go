package main

import (
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/ggoulart/michael-connelly-api/cmd/router"
	"github.com/spf13/viper"
)

func main() {
	loadConfigs()
	r := router.NewRouter()

	if viper.Get("env") == "local" {
		err := r.Run(":3000")
		if err != nil {
			log.Panic(fmt.Errorf("failed to start server: %v", err))
		}
	} else {
		adapter := ginadapter.New(r)
		lambda.Start(adapter.ProxyWithContext)
	}

}

func loadConfigs() {
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath("./configs")

	err := viper.ReadInConfig()
	if err != nil {
		log.Panic(fmt.Errorf("failed to load config file: %s", err))
	}

	viper.Set("env", "local")
}
