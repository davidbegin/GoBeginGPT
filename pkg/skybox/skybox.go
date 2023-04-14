package skybox

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"

	"beginbot.com/GoBeginGPT/pkg/utils"
)

var SKYBOX_URL = "https://backend.blockadelabs.com/api/v1/imagine"
var SKYBOX_API_KEY = os.Getenv("SKYBOX_API_KEY")

var dir, _ = utils.GetGreatGrandparentDir()
var skyboxResponseFilePath = dir + "/tmp/skybox_response.json"
var skyboxWebpageTemplateFilepath = dir + "/templates/skybox.html"
var skyboxWebpageFilepath = dir + "/build/skybox.html"

type OuterRequest struct {
	Response Response `json:"request"`
}

type Response struct {
	ID            int         `json:"id"`
	UserID        int         `json:"user_id"`
	Title         string      `json:"title"`
	Context       interface{} `json:"context"`
	Prompt        string      `json:"prompt"`
	CaptionText   interface{} `json:"caption_text"`
	AuthorName    string      `json:"author_name"`
	AliasID       interface{} `json:"alias_id"`
	AliasName     interface{} `json:"alias_name"`
	Progress      int         `json:"progress"`
	Status        string      `json:"status"`
	QueuePosition int         `json:"queue_position"`
	FileURL       string      `json:"file_url"`
	ThumbURL      string      `json:"thumb_url"`
	VideoURL      interface{} `json:"video_url"`
	CreatedAt     string      `json:"created_at"`
	UpdatedAt     string      `json:"updated_at"`
	MediaVersion  int         `json:"media_version"`
	Public        int         `json:"public"`
	ErrorMessage  interface{} `json:"error_message"`
	Type          string      `json:"type"`
	GeneratorData struct {
		Prompt        string `json:"prompt"`
		NegativeText  string `json:"negative_text"`
		AnimationMode string `json:"animation_mode"`
	} `json:"generator_data"`
	CountFavorites           int `json:"count_favorites"`
	LikesCount               int `json:"likes_count"`
	UserImaginariumImageLeft int `json:"user_imaginarium_image_left"`
}

func requestStatus(id string) Response {
	url := fmt.Sprintf("%s/requests/%s?api_key=%s", SKYBOX_URL, id, SKYBOX_API_KEY)

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var parsedResponse OuterRequest
	json.Unmarshal(body, &parsedResponse)

	fmt.Printf(
		"%s\nUpdated: %s | Status: %s\n",
		parsedResponse.Response.Prompt,
		parsedResponse.Response.UpdatedAt,
		parsedResponse.Response.Status,
	)

	return parsedResponse.Response
}

func CreateSkyboxPage(url string) {
	tmpl, err := template.ParseFiles(skyboxWebpageTemplateFilepath)
	if err != nil {
		fmt.Println("\nError Parsing Template File: ", err)
		return
	}

	f, err := os.Create(skyboxResponseFilePath)
	if err != nil {
		fmt.Printf("\nError Creating Build File: %s", err)
		return
	}

	type SkyboxPage struct {
		Url string
	}

	page := SkyboxPage{Url: url}
	err = tmpl.Execute(f, page)
	if err != nil {
		fmt.Printf("Error Executing Template File: %s", err)
		return
	}
}

func requestAll() {
	url := fmt.Sprintf("%s/myRequests?api_key=%s", SKYBOX_URL, SKYBOX_API_KEY)

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	sb := string(body)
	log.Printf(sb)
}

func parseSkyboxResponse() {
	skyboxResponse, err := os.ReadFile(skyboxResponseFilePath)
	if err != nil {
		fmt.Print("Error reading Skybox response")
		panic(err)
	}

	var parsedResponse OuterRequest
	json.Unmarshal(skyboxResponse, &parsedResponse)

	fmt.Print("\n\t ~~~ Checking Status of Skybox generation ~~~\n\n")

	id := fmt.Sprint(parsedResponse.Response.ID)
	fmt.Printf("\n\tID: %s\n", id)

	for {
		timer := time.NewTimer(5 * time.Second)
		request := requestStatus(id)

		if request.Status == "error" {
			fmt.Printf("\n\nError in skybox generation!\n\n")
			break
		}

		if request.FileURL != "" {
			fmt.Printf("Skybox URL: %s\n", request.FileURL)

			sb := request.FileURL
			d1 := []byte(sb)
			err = os.WriteFile(dir+fmt.Sprintf("/skybox_archive/%s.txt", parsedResponse.Response.Prompt[:10]), d1, 0644)
			if err != nil {
				fmt.Printf("Error Writing to skybox Archive: %+v", err)
				panic(err)
			}

			fmt.Print("Generating Skybox HTML Page!")
			CreateSkyboxPage(request.FileURL)
			break
		}

		fmt.Print("...\n\n")

		<-timer.C
	}
}

func requestImage(prompt string) {
	requestsURL := fmt.Sprintf("%s/requests?api_key=%s", SKYBOX_URL, SKYBOX_API_KEY)

	postBody, _ := json.Marshal(map[string]string{
		"prompt":    prompt,
		"generator": "stable-skybox",
		// "aspect":    "landscape",
		// "webhook_url": "https://f7a0-47-151-134-189.ngrok.io/hello",
	})
	responseBody := bytes.NewBuffer(postBody)

	resp, err := http.Post(requestsURL, "application/json", responseBody)
	if err != nil {
		log.Fatalf("An Error Occurred %v", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	sb := string(body)
	log.Printf(sb)

	d1 := []byte(sb)

	err = os.WriteFile(skyboxResponseFilePath, d1, 0644)

	if err != nil {
		fmt.Printf("%+v", err)
		panic(err)
	}
}

func GenerateSkybox(prompt string) {
	requestImage(prompt)
	parseSkyboxResponse()
}
