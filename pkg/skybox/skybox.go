package skybox

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"

	"beginbot.com/GoBeginGPT/pkg/utils"
)

var SKYBOX_REMIX_URL = "https://backend.blockadelabs.com/api/v1/skybox"
var SKYBOX_IMAGINE_URL = "https://backend.blockadelabs.com/api/v1/imagine"
var SKYBOX_API_KEY = os.Getenv("SKYBOX_API_KEY")

var dir, _ = utils.GetGreatGrandparentDir()
var skyboxResponseFilePath = dir + "/tmp/skybox_response.json"
var skyboxRemixResponseFilePath = dir + "/tmp/remix_skybox_response.json"
var skyboxWebpageTemplateFilepath = dir + "/templates/skybox.html"
var skyboxWebpageFilepath = dir + "/build/skybox.html"

type OuterRequest struct {
	Response Response `json:"request"`
}

type SkyboxStyle struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	MaxChar   string `json:"max-char"`
	Image     string `json:"image"`
	SortOrder int    `json:"sort_order"`
}

type RemixRequestResponse struct {
	ID              int       `json:"id"`
	ObfuscatedID    string    `json:"obfuscated_id"`
	UserID          int       `json:"user_id"`
	Title           string    `json:"title"`
	Prompt          string    `json:"prompt"`
	Username        string    `json:"username"`
	Status          string    `json:"status"`
	QueuePosition   int       `json:"queue_position"`
	FileURL         string    `json:"file_url"`
	ThumbURL        string    `json:"thumb_url"`
	DepthMapURL     string    `json:"depth_map_url"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	ErrorMessage    string    `json:"error_message"`
	PusherChannel   string    `json:"pusher_channel"`
	PusherEvent     string    `json:"pusher_event"`
	Type            string    `json:"type"`
	SkyboxStyleID   int       `json:"skybox_style_id"`
	SkyboxID        int       `json:"skybox_id"`
	SkyboxStyleName string    `json:"skybox_style_name"`
	SkyboxName      string    `json:"skybox_name"`
}

// "id": 5,
// "name": "Digital Painting",
// "max-char": "420",
// "image": null,
// "sort_order": 1
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

func Remix(remixID int, styleID int, prompt string) string {
	requestsURL := fmt.Sprintf("%s?api_key=%s", SKYBOX_REMIX_URL, SKYBOX_API_KEY)
	// TODO: setup logic to choose this
	// setup, using same prompt + seed + remix w/ different StyleID

	postBody, _ := json.Marshal(map[string]interface{}{
		"prompt":           prompt,
		"generator":        "stable-skybox",
		"skybox_style_id":  styleID,
		"remix_imagine_id": remixID,
	})
	responseBody := bytes.NewBuffer(postBody)
	fmt.Printf("%+v", responseBody)

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
	d1 := []byte(sb)

	// We need to save to proper remix Filepath
	err = os.WriteFile(skyboxRemixResponseFilePath, d1, 0644)

	if err != nil {
		fmt.Printf("%+v", err)
		panic(err)
	}

	// I could just pass in a bool, for Remix or not
	return ParseSkyboxRemixResponse(skyboxRemixResponseFilePath)
}

func requestStatus(id string) Response {
	url := fmt.Sprintf("%s/requests/%s?api_key=%s", SKYBOX_IMAGINE_URL, id, SKYBOX_API_KEY)

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

	f, err := os.Create(skyboxWebpageFilepath)
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
	url := fmt.Sprintf("%s/myRequests?api_key=%s", SKYBOX_IMAGINE_URL, SKYBOX_API_KEY)

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

// TODO: Don't panic, returns errors
func ParseSkyboxRemixResponse(responseFilepath string) string {
	newSkyboxURL := ""

	skyboxResponse, err := os.ReadFile(responseFilepath)
	if err != nil {
		fmt.Print("Error reading Skybox response")
		panic(err)
	}

	fmt.Printf("SKYBOX RESPONSE: %+v\n", string(skyboxResponse))

	var parsedResponse RemixRequestResponse
	json.Unmarshal(skyboxResponse, &parsedResponse)

	fmt.Print("\n\t ~~~ Checking Status of Skybox Remix generation ~~~\n\n")

	id := fmt.Sprint(parsedResponse.ID)
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

			// This is a dumb way to name your file
			err = os.WriteFile(
				dir+fmt.Sprintf(
					"/skybox_archive/%d.txt",
					parsedResponse.ID,
				),
				d1,
				0644,
			)
			if err != nil {
				fmt.Printf("Error Writing to skybox Archive: %+v", err)
				panic(err)
			}

			fmt.Printf("Generating Skybox HTML Page: %s\n", request.FileURL)
			newSkyboxURL = request.FileURL
			CreateSkyboxPage(request.FileURL)
			fmt.Print("Finished Generating HTML Page\n")
			break
		}

		fmt.Print("...\n\n")

		<-timer.C
	}

	return newSkyboxURL
}

func ParseSkyboxResponseAndUpdateWebpage(responseFilepath string) {
	skyboxResponse, err := os.ReadFile(responseFilepath)
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
			err = os.WriteFile(
				dir+fmt.Sprintf(
					"/skybox_archive/%s.txt",
					parsedResponse.Response.Prompt[:10],
				),
				d1,
				0644,
			)
			if err != nil {
				fmt.Printf("Error Writing to skybox Archive: %+v", err)
				panic(err)
			}

			fmt.Printf("Generating Skybox HTML Page: %s\n", request.FileURL)

			// we need to send a Websocket message
			// with the URL inside of it
			// this message we we will

			// CreateSkyboxPage(request.FileURL)
			// fmt.Print("Finished Generating HTML Page\n")
			// chatNotif := fmt.Sprintf("! %d | %s", parsedResponse.Response.ID, parsedResponse.Response.Prompt)
			// fmt.Printf("ChatNotif: %s\n", chatNotif)
			// notif := fmt.Sprintf("beginbot \"%s\"", chatNotif)
			// _, err := utils.RunBashCommand(notif)
			// if err != nil {
			// 	log.Fatal(err)
			// }
			break
		}

		fmt.Print("...\n\n")

		<-timer.C
	}
}

func ParseSkyboxResponseAndGenerateHTML(responseFilepath string) {
	skyboxResponse, err := os.ReadFile(responseFilepath)
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
			err = os.WriteFile(
				dir+fmt.Sprintf(
					"/skybox_archive/%s.txt",
					parsedResponse.Response.Prompt[:10],
				),
				d1,
				0644,
			)
			if err != nil {
				fmt.Printf("Error Writing to skybox Archive: %+v", err)
				panic(err)
			}

			fmt.Printf("Generating Skybox HTML Page: %s\n", request.FileURL)

			CreateSkyboxPage(request.FileURL)
			fmt.Print("Finished Generating HTML Page\n")

			chatNotif := fmt.Sprintf("! %d | %s", parsedResponse.Response.ID, parsedResponse.Response.Prompt)
			fmt.Printf("ChatNotif: %s\n", chatNotif)
			notif := fmt.Sprintf("beginbot \"%s\"", chatNotif)
			_, err := utils.RunBashCommand(notif)
			if err != nil {
				log.Fatal(err)
			}
			break
		}

		fmt.Print("...\n\n")

		<-timer.C
	}
}

func requestImage(prompt string) {
	requestsURL := fmt.Sprintf("%s/requests?api_key=%s", SKYBOX_IMAGINE_URL, SKYBOX_API_KEY)

	// before I prompt
	prompt = strings.TrimLeft(prompt, " ")
	words := strings.Split(prompt, " ")
	// fmt.Printf("Words: %+v", words)

	styleFile := dir + "/tmp/skybox_styles.json"
	body, err := ioutil.ReadFile(styleFile)
	if err != nil {
		log.Fatalln(err)
	}

	var styles []SkyboxStyle
	err = json.Unmarshal(body, &styles)
	if err != nil {
		fmt.Printf("Error parsing Skybox Styles JSON")
	}

	SkyboxStyleID := 1

	fmt.Printf("\n\nFirst Word: %s\n", words[0])

	for _, style := range styles {
		// fmt.Printf("\tSkybox Style: %d\n", style.ID)

		if fmt.Sprintf("%d", style.ID) == words[0] {
			prompt = strings.Join(words, " ")
			SkyboxStyleID = style.ID

			fmt.Printf("\tCustom Skybox Style: %s\n", style.Name)
		}
	}

	fmt.Printf("Generating Skybox w/ Custom Skybox ID: %d", SkyboxStyleID)
	postBody, _ := json.Marshal(map[string]interface{}{
		"prompt":          prompt,
		"generator":       "stable-skybox",
		"skybox_style_id": SkyboxStyleID,
		// "aspect":    "landscape",
		// "webhook_url": "https://f7a0-47-151-134-189.ngrok.io/hello",
	})
	responseBody := bytes.NewBuffer(postBody)

	resp, err := http.Post(requestsURL, "application/json", responseBody)
	if err != nil {
		log.Fatalf("An Error Occurred %v", err)
	}

	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	sb := string(body)

	d1 := []byte(sb)

	err = os.WriteFile(skyboxResponseFilePath, d1, 0644)

	if err != nil {
		fmt.Printf("%+v", err)
		panic(err)
	}
}

func RequestAllStyles() {
	baseURL := "https://backend.blockadelabs.com/api/v1/skybox/styles"

	url := fmt.Sprintf("%s?api_key=%s", baseURL, SKYBOX_API_KEY)

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	styleFile := dir + "/tmp/skybox_styles.json"

	var styles []SkyboxStyle

	err = json.Unmarshal(body, &styles)
	if err != nil {
		fmt.Printf("Error parsing Skybox Styles JSON")
	}

	err = os.WriteFile(styleFile, body, 0644)
	if err != nil {
		fmt.Printf("Error writing Skybox Styles")
	}

	var chatResponse string
	chunkCount := 0

	for i, style := range styles {
		chunkCount += 1

		if i < 1 {
			chatResponse = fmt.Sprintf("%d - %s", style.ID, style.Name)
		} else {

			if chatResponse == "" {
				// Don't include the comma is the chunkCount == 1
				chatResponse = fmt.Sprintf("%d = %s", style.ID, style.Name)
			} else {
				// Don't include the comma is the chunkCount == 1
				chatResponse = fmt.Sprintf("%s, %d = %s", chatResponse, style.ID, style.Name)
			}

			if chunkCount > 5 {
				styleOpts := fmt.Sprintf("beginbot %s", chatResponse)
				output, err := utils.RunBashCommand(styleOpts)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Printf(output)
				chatResponse = ""
				chunkCount = 0
			}

			if (i+1) == len(styles) && chunkCount < 5 {
				styleOpts := fmt.Sprintf("beginbot %s", chatResponse)
				output, err := utils.RunBashCommand(styleOpts)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Printf(output)
				chatResponse = ""
				chunkCount = 0
			}
		}
	}
}

// TODO: update this , so requestImage, passes a info to parseSkyboxResponse
// Request
func GenerateSkybox(prompt string) {
	requestImage(prompt)
	ParseSkyboxResponseAndGenerateHTML(skyboxResponseFilePath)
}
