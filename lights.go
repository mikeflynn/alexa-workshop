package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/context"
	alexa "github.com/mikeflynn/go-alexa/skillserver"
)

func EchoLights(w http.ResponseWriter, r *http.Request) {
	echoReq := context.Get(r, "echoRequest").(*alexa.EchoRequest)

	if echoReq.GetRequestType() == "LaunchRequest" {
		// Do nothing.
	} else if echoReq.GetRequestType() == "IntentRequest" {
		var echoResp *alexa.EchoResponse

		HueSetup(os.Getenv("HUE_BASESTATION_IP"), os.Getenv("HUE_BASESTATION_USER"))
		lights, err := HueGetList()

		switch echoReq.GetIntentName() {
		case "ToggleLights":
			if err != nil {
				log.Printf("Error: %v", err.Error())
				return
			}

			toggle := true

			for _, v := range lights {
				if strings.Contains(v.Name, "Workshop") {
					if v.State.On {
						toggle = false
						break
					}
				}
			}

			for k, v := range lights {
				if strings.Contains(v.Name, "Workshop") {
					for i := 0; i < 3; i++ {
						err := HueSetLight(k, HueLightState{On: toggle, Bri: 200})
						if err != nil {
							log.Printf("HUE ERROR: %v", err.Error())
						} else {
							break
						}
					}
				}
			}
		case "AllOn":
			for k, v := range lights {
				if strings.Contains(v.Name, "Workshop") {
					for i := 0; i < 3; i++ {
						err := HueSetLight(k, HueLightState{On: true, Bri: 200})
						if err != nil {
							log.Printf("HUE ERROR: %v", err.Error())
						} else {
							break
						}
					}
				}
			}
		case "AllOff":
			for k, v := range lights {
				if strings.Contains(v.Name, "Workshop") {
					for i := 0; i < 3; i++ {
						err := HueSetLight(k, HueLightState{On: false, Bri: 200})
						if err != nil {
							log.Printf("HUE ERROR: %v", err.Error())
						} else {
							break
						}
					}
				}
			}
		case "PercentOn":
			perStr, _ := echoReq.GetSlotValue("Percent")
			perInt, _ := strconv.ParseInt(perStr, 10, 8)
			for k, v := range lights {
				if strings.Contains(v.Name, "Workshop") {
					for i := 0; i < 3; i++ {
						err := HueSetLight(k, HueLightState{On: true, Bri: int(perInt)})
						if err != nil {
							log.Printf("HUE ERROR: %v", err.Error())
						} else {
							break
						}
					}
				}
			}
		case "MovieMode":
			config := map[string]interface{}{
				"front_on":  true,
				"front_bri": 100,
				"back_on":   false,
				"back_bri":  200,
			}

			ToggleWorkshopConfig(config)
		case "ComputerMode":
			config := map[string]interface{}{
				"front_on":  false,
				"front_bri": 200,
				"back_on":   true,
				"back_bri":  127,
			}

			ToggleWorkshopConfig(config)
		default:
			echoResp = alexa.NewEchoResponse().OutputSpeech("I'm sorry, I didn't get that. Can you say that again?").EndSession(false)
		}

		json, _ := echoResp.String()
		w.Header().Set("Content-Type", "application/json;charset=UTF-8")
		w.Write(json)
	} else if echoReq.GetRequestType() == "SessionEndedRequest" {
		// Do nothing.
	}
}

func ToggleWorkshopConfig(config map[string]interface{}) {
	//HueSetup(Config.Section("hue").Key("baseStationIP").String(), Config.Section("hue").Key("baseStationUser").String())

	lights, err := HueGetList()
	if err != nil {
		log.Printf("Error: %v", err.Error())
		return
	}

	inConfig := true

	// Check to see if they are in the configuration already
	for _, v := range lights {
		if strings.Contains(v.Name, "Workshop Front") {
			if v.State.On != config["front_on"].(bool) {
				inConfig = false
				break
			}
		} else if strings.Contains(v.Name, "Workshop Back") {
			if v.State.On != config["back_on"].(bool) || v.State.Bri != config["back_bri"].(int) {
				inConfig = false
				break
			}
		}
	}

	for k, v := range lights {
		if inConfig == false {
			// Turn the front lights off and the back lights down to half brightness.
			if strings.Contains(v.Name, "Workshop Front") {
				for i := 0; i < 3; i++ {
					err := HueSetLight(k, HueLightState{On: config["front_on"].(bool), Bri: config["front_bri"].(int)})
					if err != nil {
						log.Printf("HUE ERROR: %v", err.Error())
					} else {
						break
					}
				}
			} else if strings.Contains(v.Name, "Workshop Back") {
				for i := 0; i < 3; i++ {
					err := HueSetLight(k, HueLightState{On: config["back_on"].(bool), Bri: config["back_bri"].(int)})
					if err != nil {
						log.Printf("HUE ERROR: %v", err.Error())
					} else {
						break
					}
				}
			}
		} else {
			// Turn the lights back on with regular brightness.
			for i := 0; i < 3; i++ {
				err := HueSetLight(k, HueLightState{On: true, Bri: 200})
				if err != nil {
					log.Printf("HUE ERROR: %v", err.Error())
				} else {
					break
				}
			}
		}
	}
}

// Ripped from https://github.com/mikeflynn/go-dash-button/blob/master/hue.go
// ...because it's not really complete enough to put in it's own repo.

var HueBaseStationIP string
var HueUserName string

type HueLightState struct {
	Alert     string `json:"alert,omitempty"`
	Bri       int    `json:"bri,omitempty"`
	On        bool   `json:"on"`
	Reachable bool   `json:"reachable,omitempty"`
}

type HueLight struct {
	State            HueLightState `json:"state"`
	Type             string        `json:"type"`
	Name             string        `json:"name"`
	Modelid          string        `json:"modelid"`
	Manufacturername string        `json:"manufacturername"`
	Uniqueid         string        `json:"uniqueid"`
	Swversion        string        `json:"swversion"`
	Pointsymbol      struct {
		One   string `json:"1"`
		Two   string `json:"2"`
		Three string `json:"3"`
		Four  string `json:"4"`
		Five  string `json:"5"`
		Six   string `json:"6"`
		Seven string `json:"7"`
		Eight string `json:"8"`
	} `json:"pointsymbol"`
}

func HueSetup(baseStationIP string, userName string) {
	HueBaseStationIP = baseStationIP
	HueUserName = userName
}

func HueGetList() (map[string]HueLight, error) {
	response, err := http.Get("http://" + HueBaseStationIP + "/api/" + HueUserName + "/lights")
	if err != nil {
		return map[string]HueLight{}, err
	} else {
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return map[string]HueLight{}, err
		}

		// Because of Hue's weird API response we're going to unmarshal
		// to a map rather than a struct

		var lightTempMap map[string]*json.RawMessage
		err = json.Unmarshal(contents, &lightTempMap)
		if err != nil {
			return map[string]HueLight{}, err
		}

		lightMap := make(map[string]HueLight)
		for k, _ := range lightTempMap {
			var m HueLight
			_ = json.Unmarshal(*lightTempMap[k], &m)

			lightMap[k] = m
		}

		return lightMap, nil
	}
}

func HueGetLight(id string) (HueLight, error) {
	response, err := http.Get("http://" + HueBaseStationIP + "/api/" + HueUserName + "/lights/" + id)
	if err != nil {
		return HueLight{}, err
	} else {
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return HueLight{}, err
		}

		var light HueLight
		err = json.Unmarshal(contents, &light)
		if err != nil {
			return HueLight{}, err
		}

		return light, nil
	}
}

func HueSetLight(id string, options HueLightState) error {
	url := "http://" + HueBaseStationIP + "/api/" + HueUserName + "/lights/" + id + "/state"

	jsonStr, err := json.Marshal(options)
	if err != nil {
		return err
	}

	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(jsonStr))
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	contents, _ := ioutil.ReadAll(resp.Body)

	if strings.Contains(string(contents), "error") {
		return errors.New(string(contents))
	}

	return nil
}
