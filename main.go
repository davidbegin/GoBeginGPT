package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/websocket"

	"beginbot.com/GoBeginGPT/pkg/file_watchers"
	"beginbot.com/GoBeginGPT/pkg/gpt_response_parser"
	"beginbot.com/GoBeginGPT/pkg/skybox"
)

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan string)
var mutex = &sync.Mutex{}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func showAndTell(broadcast chan string) {
	done := make(chan bool)

	go file_watchers.Look4MoveRequests(broadcast)
	go file_watchers.Look4VoiceRequests(broadcast)
	go file_watchers.Look4GptRequests(broadcast)
	go file_watchers.Look4RemixRequests(broadcast)
	go file_watchers.Look4SkyboxRequests(broadcast)
	go file_watchers.Look4PreviousRequests(broadcast)
	// This is not currently working
	// go file_watchers.Look4DuetRequests(broadcast)
	go handleBroadcast()

	fmt.Println("Server is listening on :8080")
	http.HandleFunc("/ws", websocketHandler)
	http.Handle("/", http.FileServer(http.Dir("static")))
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
	<-done
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
			err := client.WriteMessage(websocket.TextMessage, []byte(content))

			if err != nil {
				fmt.Println("Error broadcasting message:", err)

				client.Close()
				delete(clients, client)

				for {
					fmt.Println("Attempting to reconnect to WebSocket...")

					url := "ws://localhost:8080"
					conn, _, err := websocket.DefaultDialer.Dial(url, nil)

					if err != nil {
						fmt.Println("Error dialing WebSocket:", err)
					} else {
						mutex.Lock()
						clients[conn] = true
						mutex.Unlock()
						fmt.Println("Successfully Connected to WebSocket.")

						err = conn.WriteMessage(websocket.TextMessage, []byte(content))
						if err != nil {
							fmt.Println("Error resending message:", err)
						}

						break
					}
				}
			}
		}

		mutex.Unlock()
	}
}

// TODO: Ponder naming all this better
func main() {
	webserver := flag.Bool("webserver", false, "Whether to run a Seal Webserver")
	duet := flag.Bool("duet", false, "Whether to run Duet code")
	skybox_styles := flag.Bool("styles", false, "Whether to query for all Skybox Styles")
	prompt_file := flag.String("prompt_file", "prompt.txt", "The file that contains the prompt")

	remix := flag.Bool("remix", false, "Whether to remix the Skybox Styles")

	var remixID int
	flag.IntVar(&remixID, "remix_id", 0, "The skybox ID of skybox you want to remix")
	var prompt string
	flag.StringVar(&prompt, "prompt", "", "The prompt you want to generate")

	var styleId int
	flag.IntVar(&styleId, "style", 20, "The Skybox Style ID")

	flag.Parse()
	fmt.Printf("INFO: %s | ID: %d\n", prompt, remixID)

	if *webserver {
		showAndTell(broadcast)
	} else if *remix {
		if prompt == "" || remixID == 0 {
			fmt.Printf("Need to pass in prompt: %s | ID: %d\n", prompt, remixID)
			return
		}
		skybox.Remix(remixID, styleId, prompt)
	} else if *skybox_styles {
		skybox.RequestAllStyles()
	} else if *duet {
		gpt_response_parser.SplitDuet(broadcast, "duet.txt")
	} else {
		b, err := os.ReadFile(*prompt_file)
		if err != nil {
			panic(fmt.Sprintf("%+v", err))
		}
		prompt := string(b)
		fmt.Printf("Prompt: %+v", prompt)
		skybox.GenerateSkybox(prompt)
	}
}
