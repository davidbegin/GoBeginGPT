package main

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"text/template"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/gorilla/websocket"
)

const (
	apiURL = "https://api.uberduck.ai/speak"
	voice  = "tacotron2" // Replace with the desired voice value
)

type SpeakResponse struct {
	URL string `json:"url"`
}

type SpeakRequest struct {
	Speech string `json:"speech"`
	Voice  string `json:"voice"`
	Text   string `json:"text"`
}

// TODO: rename
var apiKey string = os.Getenv("UBER_DUCK_KEY")
var apiSecret string = os.Getenv("UBER_DUCK_SECRET")

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }}

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan string)
var mutex = &sync.Mutex{}

type Response struct {
	StartedAt  string `json:"started_at"`
	FailedAt   string `json:"failed_at"`
	FinishedAt string `json:"finished_at"`
	Path       string `json:"path"`
}

type UberduckSpeakResponse struct {
	UUID string `json:"uuid"`
}

type OuterRequest struct {
	Request Request `json:"request"`
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

// =================================================================

func parseSkyboxRespone() {
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
		request := requestStatus(id)

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

func requestStatus(id string) Request {
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

func getGrandparentDir() (string, error) {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		return "", fmt.Errorf("failed to get caller information")
	}
	parentDir := filepath.Dir(filename)
	grandparentDir := filepath.Dir(parentDir)
	return grandparentDir, nil
}

func read_in_csv() {
	// Load the CSV file
	csvFile, err := os.Open("file_list.csv")
	if err != nil {
		log.Fatalf("Error opening the CSV file: %s", err)
	}
	defer csvFile.Close()

	// Read the CSV file
	csvReader := csv.NewReader(csvFile)
	filePaths, err := csvReader.ReadAll()
	if err != nil {
		log.Fatalf("Error reading the CSV file: %s", err)
	}

	if len(filePaths) == 0 {
		log.Fatalf("No file paths found in the CSV file.")
	}

	// Choose a random file path
	rand.Seed(time.Now().UnixNano())
	randomFilePath := filePaths[rand.Intn(len(filePaths))][0]

	// Read the content of the file
	content, err := ioutil.ReadFile(randomFilePath)
	if err != nil {
		log.Fatalf("Error reading the file at %s: %s", randomFilePath, err)
	}

	fileContent := string(content)
	fmt.Printf("Content of the file %s:\n\n%s", randomFilePath, fileContent)

}

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading to WebSocket:", err)
		return
	}

	mutex.Lock()
	clients[conn] = true
	mutex.Unlock()
}

func handleBroadcast() {
	for {
		content := <-broadcast

		mutex.Lock()
		for client := range clients {

			// So here we are writing the raw message to Websockets
			err := client.WriteMessage(websocket.TextMessage, []byte(content))
			if err != nil {
				fmt.Println("Error broadcasting message:", err)
				client.Close()
				delete(clients, client)
			}
		}
		mutex.Unlock()
	}
}

// Not sure if this is the best method still
// It also doesn't preserve single line breaks
func mdToHTML(md []byte) []byte {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return markdown.Render(doc, renderer)
}

func generateAuthHeader(apiKey, apiSecret string) string {
	auth := apiKey + ":" + apiSecret
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))
	return "Basic " + encodedAuth
}

func checkUberduckStatus(uuid string) (*Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://api.uberduck.ai/speak-status?uuid="+uuid, nil)
	if err != nil {
		panic(err)
	}

	req.SetBasicAuth(apiKey, apiSecret)
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	response := &Response{}
	err = json.Unmarshal(data, response)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Response:\n\tstarted_at: %s\n\tfailed_at: %s\n\tfinished_at: %s\n\tpath: %s\n",
		response.StartedAt, response.FailedAt, response.FinishedAt, response.Path)

	return response, nil
}

// ==================================================================================================

// This is really reading in a transcription
func textToVoice(broadcast chan string, content []byte) {
	fmt.Printf("\t=== textToVoice ===\n")
	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(apiKey+":"+apiSecret))

	dir, _ := getGrandparentDir()
	inputFile := dir + "/tmp/chatgpt_response.txt"
	// inputFile := dir + "/tmp/transcription.txt"
	outputFile := dir + "/tmp/uberduck.wav"

	contents, err := ioutil.ReadFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
		os.Exit(1)
	}

	// voice := "brock-samson"
	// HILARIOUS
	// voice := "david-bowie"
	// voice := "juice-wrld-singing"

	// voice := "donkey-kong-singing"

	// voice := "david-bowie"
	voice := "shrek"

	// voice := "sir-david-attenborough"
	// voice := "c-3po"
	// voice := "arbys"
	// voice := "e40"

	speakReq := SpeakRequest{
		Voice:  voice,
		Speech: string(contents),
	}
	jsonData, err := json.Marshal(speakReq)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)

	fmt.Println("\tRequesting to Uberduck")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	fmt.Println("Response:", string(body))
	response := &UberduckSpeakResponse{}
	err = json.Unmarshal(body, response)
	if err != nil {
		panic(err)
	}

	for {
		uberduckStatus, err := checkUberduckStatus(response.UUID)

		if err != nil {
			fmt.Println("Error:", err)
		} else if uberduckStatus.FailedAt != "" {
			fmt.Println("Failed! ", err)
			break
		} else if uberduckStatus.FinishedAt != "" {

			resp, err = http.Get(uberduckStatus.Path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error downloading the WAV file: %v | %s\n", err, uberduckStatus.Path)
				os.Exit(1)
			}
			defer resp.Body.Close()

			// Save the WAV file to the output file path
			data, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading WAV data: %v\n", err)
				os.Exit(1)
			}
			err = ioutil.WriteFile(outputFile, data, 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error writing the output WAV file: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("WAV file saved successfully to '%s'\n", outputFile)

			// Does this actually update shit???
			broadcast <- string(contents)

			// This plays the Audio
			Talk()
			break
		}
		time.Sleep(3 * time.Second)
	}
}

func Talk() string {
	dir, _ := getGrandparentDir()
	talk := dir + "/talk.sh"
	cmd, err := exec.Command("/bin/sh", talk).Output()
	if err != nil {
		fmt.Printf("error %s", err)
	}
	output := string(cmd)
	return output
}

// ========================================================================================================
// ========================================================================================================
// ========================================================================================================

func showAndTell() {
	fmt.Print("\tStarting Up\n")
	dir, _ := getGrandparentDir()

	// So we just need to change file
	// file := dir + "/tmp/transcription.txt"
	file := dir + "/tmp/chatgpt_response.txt"
	fc, err := os.ReadFile(file)
	if err != nil {
		fmt.Print("Error reading contents")
	}
	fileContents := string(fc)

	fmt.Printf("\t === Watching Response File: %s\n", file)
	http.HandleFunc("/ws", websocketHandler)

	http.Handle("/", http.FileServer(http.Dir("static")))

	// The problem is we aren't watching the other files:
	// -> Mainly Transcription
	go func() {
		ticker := time.NewTicker(1 * time.Second)

		for range ticker.C {
			contents, err := os.ReadFile(file)

			if err != nil {
				fmt.Print("Error reading contents")
				continue
			}

			if string(contents) != fileContents {
				fileContents = string(contents)

				md := []byte(fileContents)
				html := mdToHTML(md)

				fmt.Print("Converting MD to HTML and broadcasting\n")

				// This calls Uberduck
				textToVoice(broadcast, html)

				broadcast <- "done"

			}
		}
	}()

	go handleBroadcast()

	fmt.Println("Server is listening on :8080")
	err = http.ListenAndServe(":8080", nil)

	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func main() {
	// TODO: Ponder naming all this better
	webserver := flag.Bool("webserver", false, "Whether to run a Seal Webserver")
	prompt_file := flag.String("prompt_file", "prompt.txt", "The file that contains the prompt")

	flag.Parse()
	if *webserver {
		showAndTell()
	} else {
		b, err := os.ReadFile(*prompt_file)
		if err != nil {
			panic(fmt.Sprintf("%+v", err))
		}
		prompt := string(b)
		fmt.Printf("Prompt: %+v", prompt)
		generateSkybox(prompt)
	}
}

func generateSkybox(prompt string) {
	requestImage(prompt)
	parseSkyboxRespone()
}
