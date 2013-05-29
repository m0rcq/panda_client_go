package main

import (
	"flag"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/m0rcq/panda_client_go"
	"log"
)

const (
	defaultCloudId   = "<INSERT YOUR CLOUD ID HERE>"
	defaultAccessKey = "<INSERT YOUR ACCESS KEY HERE>"
	defaultSecretKey = "<INSERT YOUR SECRET KEY HERE>"
	defaultApiHost   = "api.pandastream.com"
)

var cloudId string
var accessKey string
var secretKey string
var apiHost string

var pandaResource string
var pandaCmd string
var pandaObjId string
var videoFile string

func init() {
	flag.StringVar(&cloudId, "cloudId", defaultCloudId, "Set Panda Cloud ID")
	flag.StringVar(&accessKey, "accessKey", defaultAccessKey, "Set Panda Access Key")
	flag.StringVar(&secretKey, "secretKey", defaultSecretKey, "Set Panda Secret Key")
	flag.StringVar(&apiHost, "apiHost", defaultApiHost, "Set Panda API host")

	flag.StringVar(&pandaResource, "resource", "Get", "Set Panda resource")
	flag.StringVar(&pandaCmd, "cmd", "Get", "Set Panda command")
	flag.StringVar(&pandaObjId, "id", "0", "Set Panda object ID")
	flag.StringVar(&videoFile, "video", "0", "Set Video file")
}

func main() {
	flag.Parse()

	var client panda.PandaApiInterface = &panda.PandaApi{}

	client.Init(accessKey, secretKey, cloudId, apiHost, 80)

	switch {
	case pandaResource == "videos":
		switch {
		case pandaCmd == "info":
			resp, err := client.Get("/videos.json", nil)
			if err != nil {
				log.Fatal(err.Error())
			}

			js, err := simplejson.NewJson([]byte(resp))
			if err != nil {
				log.Fatal(err.Error())
			}

			videos, err := js.Array()

			if err != nil {
				log.Fatal(err.Error())
			}

			for _, v := range videos {
				video := v.(map[string]interface{})

				// access by name
				fmt.Println("Video ID:", video["id"].(string))
				fmt.Println("Original Filename:", video["original_filename"].(string))
				fmt.Println("Status:", video["status"].(string))

				error_message := "n/a"

				if video["error_message"] != nil {
					error_message = video["error_message"].(string)
				}

				fmt.Println("Error Message:", error_message)
				fmt.Println("============")
			}

		case pandaCmd == "delete":
			resp, err := client.Delete("/videos/"+pandaObjId+".json", nil)
			if err != nil {
				log.Fatal(err.Error())
			}

			fmt.Println(resp)

		case pandaCmd == "upload":
			resp, err := client.Post("/videos.json", map[string]string{"file": videoFile, "payload": "1234"})

			if err != nil {
				log.Fatal(err.Error())
			}

			js, err := simplejson.NewJson([]byte(resp))
			if err != nil {
				log.Fatal(err.Error())
			}

			if id, ok := js.CheckGet("id"); ok {
				id_str, _ := id.String()
				fmt.Println("Video ID: ", id_str)
			}

			if fn, ok := js.CheckGet("original_filename"); ok {
				fn_str, _ := fn.String()
				fmt.Println("Original filename: ", fn_str)
			}

		}
	case pandaResource == "encodings":
		switch {
		case pandaCmd == "info":
			encodings, err := client.Get("/encodings.json", map[string]string{"status": "success"})
			if err != nil {
				log.Fatal(err.Error())
			}

			fmt.Println(encodings)
		default:
			log.Fatal("Video Command not recognised")
		}
	}
}
