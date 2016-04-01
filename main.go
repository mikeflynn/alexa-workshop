package main

import (
	"math/rand"
	"os"
	"time"

	alexa "github.com/mikeflynn/go-alexa/skillserver"
)

var Applications = map[string]interface{}{
	"/echo/jeopardy": alexa.EchoApplication{
		AppID:   os.Getenv("JEOPARDY_APP_ID"),
		Handler: EchoJeopardy,
	},
	"/echo/lights": alexa.EchoApplication{
		AppID:   os.Getenv("LIGHTS_APP_ID"),
		Handler: EchoLights,
	},
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	alexa.Run(Applications, "3000")
}
