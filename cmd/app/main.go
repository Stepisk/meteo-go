package main

import "gitlab.com/peleng-meteo/meteo-go/internal/app"

const configsDir = "configs"

func main() {
	app.Run(configsDir)
}
