package file_watchers

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"beginbot.com/GoBeginGPT/pkg/skybox"
	"beginbot.com/GoBeginGPT/pkg/uberduck"
	"beginbot.com/GoBeginGPT/pkg/utils"
)

var dir, _ = utils.GetGreatGreatGrandparentDir()
var duetResp = dir + "/tmp/current/duet.txt"
var voiceCharacterFile = dir + "/tmp/voice_character.csv"
var moveRequest = dir + "/tmp/current/move.txt"
var remixRequestPath = dir + "/tmp/current/remix.txt"
var skyboxRequestPath = dir + "/tmp/current/skybox.txt"
var previousRequestPath = dir + "/tmp/current/previous.txt"
var gptResp = dir + "/tmp/current/chatgpt_response.txt"
var voiceLoc = dir + "/tmp/current/voice.txt"

var DEFAULT_STYLE_ID = 5

func Look4PreviousRequests(broadcast chan string) {
	done := make(chan bool)

	ogPreviousRequest, err := ioutil.ReadFile(previousRequestPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading OG GPT Response file: %v\n", err)
		os.Exit(1)
	}

	go func() {
		oneSec := time.NewTicker(1 * time.Second)

		for {
			previous, err := ioutil.ReadFile(previousRequestPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading OG GPT Response file: %v\n", err)
			}

			if string(ogPreviousRequest) != string(previous) {
				fmt.Printf("Attempting to return to previous Skybox: %s", string(previous))
				var archivePath = dir + fmt.Sprintf("/GoBeginGPT/skybox_archive/%s.txt", previous)
				// previous
				previousURL, err := ioutil.ReadFile(archivePath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error reading Archive file: %v\n", err)
					ogPreviousRequest = previous
					return
				}

				broadcast <- fmt.Sprintf("skybox %s", previousURL)
				// Read in the Skybox ID form thje file
				// look up the skybox file in the archive by ID
				// read in the contents of the skybox, which returns URL
				// send a skybox update message to websockets URL
				ogPreviousRequest = previous
			}
			<-oneSec.C
		}
	}()

	// not sure I wanna block forever here, or if I need to with the GoRoutine Above
	<-done
}
func Look4DuetRequests(broadcast chan string) {
	done := make(chan bool)

	ogGPT, err := ioutil.ReadFile(duetResp)
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

			_, err = ioutil.ReadFile(voiceLoc)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading voice file: %v\n", err)
				os.Exit(1)
			}

			if string(ogGPT) != string(gpt) {
				ogGPT = gpt
				// gpt_response_parser.SplitDuet(broadcast, "duet.txt")
			}
		}
	}()

	// not sure I wanna block forever here, or if I need to with the GoRoutine Above
	<-done
}

func Look4VoiceRequests(broadcast chan string) {
	ticker := time.NewTicker(1 * time.Second)

	for {
		voice_data := uberduck.ReadAndTruncateVoiceFile(voiceCharacterFile)
		if len(voice_data) > 0 {

			for i, vd := range voice_data {
				if i > 0 {
					character := uberduck.RandomVoiceToCharacter(vd.Voice)
					if string(vd.Text[0]) != "!" {
						// Check an overwrite file
						// OverVoice
						// This needs to take in broadcast
						uberduck.TextToVoice(broadcast, character, vd.Voice, vd.Text)
					}

					broadcast <- fmt.Sprintf("done %s", character)
				}
			}

		}
		<-ticker.C
	}
}

func Look4SkyboxRequests(broadcast chan string) {
	requestPath := skyboxRequestPath

	done := make(chan bool)

	// Load the last skybox request, before checking for updates
	ogSkybox, err := ioutil.ReadFile(requestPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading OG GPT Response file: %v\n", err)
		os.Exit(1)
	}

	go func() {
		timer := time.NewTicker(200 * time.Millisecond)

		for {
			skyboxRequest, err := ioutil.ReadFile(requestPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading OG GPT Response file: %v\n", err)
				os.Exit(1)
			}

			if string(ogSkybox) != string(skyboxRequest) {
				fmt.Printf("\tNew Skybox Request: %s", string(skyboxRequest))

				id, url := skybox.GenerateSkybox(string(skyboxRequest))

				chatNotif := fmt.Sprintf(
					"! %d | %s",
					id,
					string(skyboxRequest),
				)
				fmt.Printf("ChatNotif: %s\n", chatNotif)
				notif := fmt.Sprintf("beginbot \"%s\"", chatNotif)
				_, err := utils.RunBashCommand(notif)
				if err != nil {
					log.Fatal(err)
				}

				// We also need to send a message back to github
				broadcast <- fmt.Sprintf("skybox %s", url)

				ogSkybox = skyboxRequest
			}
			<-timer.C
		}
	}()

	// not sure I wanna block forever here, or if I need to with the GoRoutine Above
	<-done
}

// Filename to monitor
// how long to wait in-between
// take websockets connection
func Look4RemixRequests(broadcast chan string) {
	done := make(chan bool)

	ogRemix, err := ioutil.ReadFile(remixRequestPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading OG GPT Response file: %v\n", err)
		os.Exit(1)
	}

	go func() {
		oneSec := time.NewTicker(200 * time.Millisecond)
		for {
			<-oneSec.C
			remixRequest, err := ioutil.ReadFile(remixRequestPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading OG GPT Response file: %v\n", err)
				os.Exit(1)
			}

			if string(ogRemix) != string(remixRequest) {
				fmt.Printf("We are trying to move you to: %s", string(remixRequest))
				// We need to do do all the Remix shit!!!!!!

				remixID, styleId, prompt := parseRemixRequest(string(remixRequest))
				skyboxURL := skybox.Remix(remixID, styleId, prompt)
				broadcast <- fmt.Sprintf("skybox %s", skyboxURL)
				ogRemix = remixRequest
			}
		}
	}()

	// not sure I wanna block forever here, or if I need to with the GoRoutine Above
	<-done
}

// !remix REMIX_ID POTENTIAL_STYLE_ID POTENTIAL_PROMPT
func parseRemixRequest(content string) (int, int, string) {
	splitmsg := strings.Fields(content)

	defaultRemixID := 2443168
	defaultPrompt := "danker"

	remixIdStr := fmt.Sprintf("%d", defaultRemixID)
	if len(splitmsg) > 0 {
		remixIdStr = splitmsg[0]
	}
	remixID, err := strconv.Atoi(remixIdStr)
	if err != nil {
		remixID = defaultRemixID
	}

	styleIDStr := defaultPrompt
	if len(splitmsg) > 1 {
		styleIDStr = splitmsg[1]
	}
	styleID, err := strconv.Atoi(styleIDStr)

	if err != nil {
		styleID = DEFAULT_STYLE_ID
	}

	promptSkip := 1
	if styleID != 0 {
		promptSkip = 2
	}

	prompt := strings.Join(splitmsg[promptSkip:], " ")
	// fmt.Println("remixID:", remixID)
	// fmt.Println("styleID:", styleID)
	// fmt.Println("prompt:", prompt)
	return remixID, styleID, prompt
}

func Look4MoveRequests(broadcast chan string) {
	done := make(chan bool)

	ogPosition, err := ioutil.ReadFile(moveRequest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading OG GPT Response file: %v\n", err)
		os.Exit(1)
	}

	go func() {
		oneSec := time.NewTicker(200 * time.Millisecond)
		for {
			<-oneSec.C
			position, err := ioutil.ReadFile(moveRequest)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading OG GPT Response file: %v\n", err)
				os.Exit(1)
			}

			if string(ogPosition) != string(position) {
				fmt.Printf("We are trying to move you to: %s", string(position))
				broadcast <- string(position)
				ogPosition = position
			}
		}
	}()

	// not sure I wanna block forever here, or if I need to with the GoRoutine Above
	<-done
}

func Look4GptRequests(broadcast chan string) {
	done := make(chan bool)

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

			_, err = ioutil.ReadFile(voiceLoc)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading voice file: %v\n", err)
				os.Exit(1)
			}

			if string(ogGPT) != string(gpt) {
				uberduck.TextToVoice(broadcast, "birb", "neo", string(gpt))
				ogGPT = gpt
			}
		}
	}()

	// not sure I wanna block forever here, or if I need to with the GoRoutine Above
	<-done
}
