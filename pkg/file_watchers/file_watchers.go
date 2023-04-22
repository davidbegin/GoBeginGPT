package file_watchers

import (
	"fmt"
	"io/ioutil"
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
var gptResp = dir + "/tmp/current/chatgpt_response.txt"
var voiceLoc = dir + "/tmp/current/voice.txt"

var DEFAULT_STYLE_ID = 5

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
	done := make(chan bool)

	ogSkybox, err := ioutil.ReadFile(skyboxRequestPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading OG GPT Response file: %v\n", err)
		os.Exit(1)
	}

	go func() {
		oneSec := time.NewTicker(200 * time.Millisecond)
		for {
			<-oneSec.C
			skyboxRequest, err := ioutil.ReadFile(skyboxRequestPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading OG GPT Response file: %v\n", err)
				os.Exit(1)
			}

			if string(ogSkybox) != string(skyboxRequest) {
				fmt.Printf("We are trying to move you to: %s", string(skyboxRequest))
				skyboxURL := skybox.ParseSkyboxResponseAndUpdateWebpage()
				broadcast <- fmt.Sprintf("skybox %s", skyboxURL)
				ogSkybox = skyboxRequest
			}
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
