package uberduck

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"time"

	"beginbot.com/GoBeginGPT/pkg/utils"
)

var dir, _ = utils.GetGreatGreatGrandparentDir()

// TODO: rename
var apiKey string = os.Getenv("UBER_DUCK_KEY")
var apiSecret string = os.Getenv("UBER_DUCK_SECRET")

var voicesFolder = dir + "/tmp/voices"

// TODO: Abstract this
var animateScript = "/home/begin/code/BeginGPT/animate_scripts/animate.sh"

type UberduckSpeakResponse struct {
	UUID string `json:"uuid"`
}

const apiURL = "https://api.uberduck.ai/speak"

type VoiceText struct {
	Voice string
	Text  string
}

type SpeakRequest struct {
	Speech string `json:"speech"`
	Voice  string `json:"voice"`
	Text   string `json:"text"`
}

type UberduckResponse struct {
	StartedAt  string `json:"started_at"`
	FailedAt   string `json:"failed_at"`
	FinishedAt string `json:"finished_at"`
	Path       string `json:"path"`
	Detail     string `json:"detail"`
}

// =================================================================================

func TextToVoice(
	broadcast chan string,
	character string,
	voice string,
	contents string,
) {
	response := requestImageFromUberduck(character, voice, contents)

	for {
		filename := fmt.Sprintf("%s.wav", voice)
		outputFile, err := pollForFinishedUberduck(voice, filename, response.UUID)

		if err != nil {
			fmt.Println("Error from Uberduck: ", err)
			break
		}

		if outputFile != "" {
			md := []byte(contents)
			html := utils.MdToHTML(md)
			broadcast <- fmt.Sprintf("dialog %s %s", character, string(html))
			utils.Talk(outputFile)
			break
		}

		// Sleep 3 seconds and call uberduck again!
		time.Sleep(3 * time.Second)
	}
}

func TextToVoiceAndAnimate(
	broadcast chan string,
	character string,
	voice string,
	voiceFile string,
	animationNamespace string,
	contents string,
) {

	response := requestImageFromUberduck(character, voice, contents)
	dialogueFile := voicesFolder + fmt.Sprintf("/%s/%s.txt", animationNamespace, voiceFile)

	for {
		wavFile := fmt.Sprintf("%s/%s.wav", animationNamespace, voiceFile)
		outputFile, err := pollForFinishedUberduck(voice, wavFile, response.UUID)

		if err != nil {
			fmt.Println("Error from Uberduck: ", err)
			break
		}

		if outputFile != "" {
			// I don't need .wav
			// dialogueFile := voicesFolder + fmt.Sprintf("/%s", filename)
			err = ioutil.WriteFile(dialogueFile, []byte(contents), 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error writing the output file: %v\n", err)
				return
			}

			fmt.Printf("\n\t~~~~ Starting Animation: \n\t%s\n", animateScript)

			simulatedCommand := fmt.Sprintf("animate_scripts/animate.sh %s %s %s\n",
				voiceFile,
				dialogueFile,
				animationNamespace,
			)
			fmt.Printf("\n%s\n\n", simulatedCommand)

			cmd := exec.Command("/bin/sh", animateScript, voiceFile, dialogueFile, animationNamespace)
			var outb, errb bytes.Buffer
			cmd.Stdout = &outb
			cmd.Stderr = &errb
			err = cmd.Run()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Animation out:", outb.String(), "err:", errb.String())

			// So this broadcast isn't working????
			// We need to not always call this
			// broadcast <- fmt.Sprintf("start_animation")

			break
			// This leaves the function right???
		}
		time.Sleep(3 * time.Second)
	}
}

func requestImageFromUberduck(
	character string,
	voice string,
	contents string,
) *UberduckSpeakResponse {
	speakReq := SpeakRequest{
		Voice:  voice,
		Speech: contents,
	}
	response := &UberduckSpeakResponse{}

	jsonData, err := json.Marshal(speakReq)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return response
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return response
	}

	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(apiKey+":"+apiSecret))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return response
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return response
	}

	fmt.Println("\n\tUberduck Response:", string(body))

	err = json.Unmarshal(body, response)
	if err != nil {
		panic(err)
	}

	return response
}

// This needs a namespace
func pollForFinishedUberduck(voice string, filename string, uuid string) (string, error) {
	uberduckStatus, err := checkUberduckStatus(voice, uuid)
	outputFile := ""

	if err != nil || uberduckStatus.StartedAt == "" || uberduckStatus.FailedAt != "" {
		return outputFile, err
	} else if uberduckStatus.FinishedAt != "" {
		resp, err := http.Get(uberduckStatus.Path)
		if err != nil {
			fmt.Fprintf(
				os.Stderr,
				"Error downloading the WAV file: %v | %s\n",
				err,
				uberduckStatus.Path,
			)
			os.Exit(1)
		}
		defer resp.Body.Close()

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading WAV data: %v\n", err)
			os.Exit(1)
		}

		outputFile := voicesFolder + fmt.Sprintf("/%s", filename)
		err = ioutil.WriteFile(outputFile, data, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing the output WAV file: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\t~ New Audio File: '%s'\n", outputFile)
		return outputFile, nil
	}

	return "", nil
}

// =========================================================
// =========================================================
// =========================================================

func ReadAndTruncateVoiceFile(filePath string) []VoiceText {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	var voiceTexts []VoiceText
	// Process the CSV data
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		voiceTexts = append(voiceTexts, VoiceText{
			Voice: record[0],
			Text:  record[1],
		})
	}
	err = os.Truncate(filePath, 0)
	if err != nil {
		fmt.Println("Error truncating file:", err)
		return voiceTexts
	}
	return voiceTexts
}

func RandomVoiceToCharacter(voice string) string {
	choices := []string{
		"snitch",
		"crabs",
		"kevin",
		"mcguirk",
		"birb",
	}

	i := rand.Intn(len(choices))
	choice := choices[i]
	return choice
}

// =============================================================================
// =============================================================================
// =============================================================================

func checkUberduckStatus(voice string, uuid string) (*UberduckResponse, error) {
	client := &http.Client{}
	response := &UberduckResponse{}
	req, err := http.NewRequest("GET", "https://api.uberduck.ai/speak-status?uuid="+uuid, nil)
	if err != nil {
		panic(err)
	}

	req.SetBasicAuth(apiKey, apiSecret)
	resp, err := client.Do(req)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return response, err
	}

	err = json.Unmarshal(data, response)
	if err != nil {
		return response, err
	}

	return response, nil
}
