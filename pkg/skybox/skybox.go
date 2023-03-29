package skybox

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"text/template"
	"time"
)

func main() {
	fmt.Println("vim-go")
}

func getGrandparentDir() (string, error) {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		return "", fmt.Errorf("failed to get caller information")
	}
	parentDir := filepath.Dir(filename)
	grandparentDir := filepath.Dir(parentDir)
	return grandparentDir, nil
}

type OuterRequest struct {
	Request Request `json:"request"`
}

func ParseSkyboxRespone() {
	dir, err := getGrandparentDir()
	file := dir + "/tmp/skybox_response.json"
	skyboxResponse, err := os.ReadFile(file)
	if err != nil {
		fmt.Print("Error reading Skybox response")
		panic(err)
	}

	var parsedResponse OuterRequest
	json.Unmarshal(skyboxResponse, &parsedResponse)

	fmt.Print("\n\t ~~~ Checking Status of Skybox generation ~~~\n\n")

	id := fmt.Sprint(parsedResponse.Request.ID)
	fmt.Printf("\n\tID: %s", id)

	for {
		timer := time.NewTimer(5 * time.Second)
		request := RequestStatus(id)

		if request.Status == "error" {
			fmt.Printf("\n\nError in skybox generation!\n\n")
			break
		}

		// Go lang
		// So Where do I save this????
		// request.Prompt
		// SO this FILE URL
		// We need to save it, with the prompt
		if request.FileURL != "" {
			fmt.Printf("Skybox URL: %s", request.FileURL)

			// So we don't have the URL HERE!!!
			// Save URL in the Archive
			sb := request.FileURL
			d1 := []byte(sb)
			err = os.WriteFile(dir+fmt.Sprintf("/skybox_archive/%s.txt", parsedResponse.Request.Prompt[:10]), d1, 0644)
			if err != nil {
				fmt.Printf("%+v", err)
				panic(err)
			}

			fmt.Print("Generating Skybox HTML Page!")
			CreateSkyboxPage(request.FileURL)
			break
		}

		fmt.Print("\t...Waiting 5 seconds before checking again\n")

		<-timer.C
	}
}

func RequestStatus(id string) Request {
	api_key := os.Getenv("SKYBOX_API_KEY")
	url := fmt.Sprintf("https://backend.blockadelabs.com/api/v1/imagine/requests/%s?api_key=%s", id, api_key)

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

	var parsedResponse OuterRequest
	json.Unmarshal(body, &parsedResponse)

	return parsedResponse.Request
}

func CreateSkyboxPage(url string) {
	absPath, err := filepath.Abs(os.Args[0])
	if err != nil {
		panic(err)
	}
	dirPath := filepath.Dir(absPath)

	tmpl, err := template.ParseFiles(dirPath + "/templates/skybox.html")
	if err != nil {
		fmt.Println("\nError Parsing Template File: ", err)
		return
	}

	buildFile := dirPath + "/build/skybox.html"
	f, err := os.Create(buildFile)
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

type Request struct {
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

func GenerateSkybox(prompt string) {
	requestImage(prompt)
	ParseSkyboxRespone()
}

func requestImage(prompt string) {
	api_key := os.Getenv("SKYBOX_API_KEY")

	// TODO: extract out base URL
	// webhook_url := "https://bc26-47-151-134-189.ngrok.io/hello"
	requestsURL := fmt.Sprintf("https://backend.blockadelabs.com/api/v1/imagine/requests?api_key=%s", api_key)

	// b, err := os.ReadFile(*prompt_file)
	// prompt := string(b)
	// fmt.Printf("%+v", prompt)

	postBody, _ := json.Marshal(map[string]string{
		"prompt":    prompt,
		"generator": "stable-skybox",
		// "api_key":   api_key,
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

	dir, err := getGrandparentDir()

	// We need to make sure this is consistent
	err = os.WriteFile(dir+"/tmp/skybox_response.json", d1, 0644)

	if err != nil {
		fmt.Printf("%+v", err)
		panic(err)
	}
}

func requestAll() {
	api_key := os.Getenv("SKYBOX_API_KEY")

	url := fmt.Sprintf("https://backend.blockadelabs.com/api/v1/imagine/myRequests?api_key=%s", api_key)

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
