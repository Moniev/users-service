package main

import "users-service/app/config"

func main() {
	settings := config.GetSettings()

	router := config.InitApp(settings)
	router.Run(settings.ApiPort)
}
