package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"beginbot.com/GoBeginGPT/pkg/gpt_response_parser"
	"beginbot.com/GoBeginGPT/pkg/skybox"
	"beginbot.com/GoBeginGPT/pkg/uberduck"
	"beginbot.com/GoBeginGPT/pkg/utils"
)

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan string)
var mutex = &sync.Mutex{}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func look4GptRequests() {
	done := make(chan bool)

	dir, _ := utils.GetGrandparentDir()
	gptResp := dir + "/tmp/current/chatgpt_response.txt"
	ogGPT, err := ioutil.ReadFile(gptResp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading OG GPT Response file: %v\n", err)
		os.Exit(1)
	}

	go func() {
		oneSec := time.NewTicker(1 * time.Second)
		for {
			<-oneSec.C
			gpt, err := ioutil.ReadFile(gptResp)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading OG GPT Response file: %v\n", err)
				os.Exit(1)
			}

			voiceLoc := dir + "/tmp/current/voice.txt"
			_, err = ioutil.ReadFile(voiceLoc)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading voice file: %v\n", err)
				os.Exit(1)
			}

			// WE need some way knowing it's a duet
			if string(ogGPT) != string(gpt) {
				gpt_response_parser.SplitDuet("chatgpt_response.txt")
			}
		}
	}()

	// not sure I wanna block forever here, or if I need to with the GoRoutine Above
	<-done
}

func look4VoiceRequests() {
	dir, _ := utils.GetGrandparentDir()

	// This might still be right
	// MARK MARK
	voiceCharacterFile := dir + "/tmp/voice_character.csv"
	ticker := time.NewTicker(1 * time.Second)
	for {
		voice_data := uberduck.ReadAndTruncateVoiceFile(voiceCharacterFile)
		if len(voice_data) > 0 {

			for i, vd := range voice_data {
				if i > 0 {
					character := uberduck.RandomVoiceToCharacter(vd.Voice)
					if string(vd.Text[0]) != "!" {

						// Check an overwrite file
						// overVoice
						// This needs to take in broadcast
						uberduck.TextToVoice(character, vd.Voice, vd.Text)
					}
					broadcast <- fmt.Sprintf("done %s", character)
				}
			}

		}
		<-ticker.C
	}
}

// This is reall the Webserver that runs
func showAndTell() {
	done := make(chan bool)

	// I could also pass done to each of these to wait
	go look4VoiceRequests()

	go look4GptRequests()

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

		// qwertywert_: Hey is it because on the retry, you add back to clients[], so outer keeps looping back in?
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
						fmt.Println("Successfully reconnected to WebSocket.")

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

func main() {
	// TODO: Ponder naming all this better
	webserver := flag.Bool("webserver", false, "Whether to run a Seal Webserver")
	duet := flag.Bool("duet", false, "Whether to run Duet code")
	prompt_file := flag.String("prompt_file", "prompt.txt", "The file that contains the prompt")

	flag.Parse()
	if *webserver {
		showAndTell()
	} else if *duet {
		gpt_response_parser.SplitScript()
		gpt_response_parser.SplitDuet("chatgpt_response.txt")
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
