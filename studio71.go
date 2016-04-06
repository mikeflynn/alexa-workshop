package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	alexa "github.com/mikeflynn/go-alexa/skillserver"
)

func EchoStudio71(echoReq *alexa.EchoRequest, echoResp *alexa.EchoResponse) {
	switch echoReq.GetIntentName() {
	case "Status":
		res, err := http.Get("http://status.studio71.io/status.json")
		if err != nil {
			log.Printf("Error fetching status json: %v", err.Error())
		}

		defer res.Body.Close()

		decoder := json.NewDecoder(res.Body)
		var data S71Status
		err = decoder.Decode(&data)
		if err != nil {
			log.Printf("Error parsing status json: %v", err.Error())
		}

		message := ""
		for _, app := range data.Applications {
			if app.Status != "UP" {
				message += fmt.Sprintf("%v is down.", app.Name)
			}
		}

		if message == "" {
			message = getRandom([]string{
				"All applications are up and running normally.",
				"Everything is good to go.",
				"All sites are running.",
				"Everything is normal.",
			})
		}

		echoResp.OutputSpeech(message).EndSession(true)
	default:
		echoResp.OutputSpeech("I'm sorry, I didn't get that. Can you say that again?").EndSession(false)
	}
}

type S71Status struct {
	Metadata []struct {
		Timestamp string `json:"timestamp,omitempty"`
		Host      string `json:"host,omitempty"`
	} `json:"metadata"`
	Applications []struct {
		Name      string      `json:"name"`
		Status    string      `json:"status"`
		DownSince interface{} `json:"down_since"`
	} `json:"applications"`
	FetchServers []struct {
		Name      string      `json:"name"`
		Status    string      `json:"status"`
		IP        string      `json:"ip"`
		DownSince interface{} `json:"down_since"`
	} `json:"fetch_servers"`
}
