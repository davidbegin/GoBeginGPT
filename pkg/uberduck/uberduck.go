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
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

var broadcast = make(chan string)

// TODO: rename
var apiKey string = os.Getenv("UBER_DUCK_KEY")
var apiSecret string = os.Getenv("UBER_DUCK_SECRET")

type UberduckSpeakResponse struct {
	UUID string `json:"uuid"`
}

const (
	apiURL = "https://api.uberduck.ai/speak"
)

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

func TextToVoice(character string, voice string, contents string) {
	dir, _ := utils.GetGrandparentDir()

	speakReq := SpeakRequest{
		Voice:  voice,
		Speech: contents,
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

	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(apiKey+":"+apiSecret))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)

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

	fmt.Println("\n\tUberduck Response:", string(body))

	response := &UberduckSpeakResponse{}
	err = json.Unmarshal(body, response)
	if err != nil {
		panic(err)
	}

	for {
		uberduckStatus, err := checkUberduckStatus(voice, response.UUID)

		if err != nil {
			fmt.Println("Error:", err)
			break
		} else if uberduckStatus.StartedAt == "" {
			fmt.Println("Failed! ", err)
			break
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

			// TODO: Look at this more!!!!!!!!
			outputFile := dir + fmt.Sprintf("/../../tmp/voices/%s.wav", voice)
			err = ioutil.WriteFile(outputFile, data, 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error writing the output WAV file: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("\t~ New Audio File: '%s'\n", outputFile)

			md := []byte(contents)
			html := mdToHTML(md)
			// Does this actually update shit???
			// this is the actual text
			// which means we need the

			// We need to save this character
			broadcast <- fmt.Sprintf("dialog %s %s", character, string(html))

			// This plays the Audio
			Talk(outputFile)
			break
		}
		time.Sleep(3 * time.Second)
	}
}

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

func Talk(soundFile string) string {
	// So this doesn't spawn a new process
	soundCmd := fmt.Sprintf("play %s", soundFile)
	output, err := runBashCommand(soundCmd)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(output)
	return output
}

// This & didn't do anything
func runBashCommand(command string) (string, error) {
	cmd := exec.Command("bash", "-c", command, "&")
	outputBytes, err := cmd.Output()
	if err != nil {
		return "", err
	}
	output := string(outputBytes)
	return output, nil
}

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

// We should pass the:
//   - name of voicefile we want to save
func TextToVoiceAndAnimate(
	character string,
	voice string,
	voiceFile string,
	animationNamespace string,
	contents string,
) {
	dir, _ := utils.GetGrandparentDir()

	// fullScript := splitScript()
	fullScript := contents
	// Take the contents
	// Split into prompt
	// and audioScript
	// generateSkybox(prompt)
	// We need to parse Skybox Requests

	// Before all this,
	// We need to parse out images to generate
	// Send them to Skybox
	// md := []byte(contents)
	// text := mdToHTML(md)

	speakReq := SpeakRequest{
		Voice:  voice,
		Speech: fullScript,
		// Text:   string(text),
	}

	jsonData, err := json.Marshal(speakReq)
	if err != nil {
		fmt.Printf("Error Marshaling speakReq: %s\n", err)
		return
	}

	fmt.Printf("jsonData: %+v\n", string(jsonData))
	client := &http.Client{}
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error Creating request: %s\n", err)
		return
	}
	authHeader := "Basic " + base64.StdEncoding.EncodeToString(
		[]byte(apiKey+":"+apiSecret),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error From Uberduck: %s\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error Reading Response Body: %s\n", err)
		return
	}

	response := &UberduckSpeakResponse{}
	err = json.Unmarshal(body, response)
	if err != nil {
		panic(err)
	}

	for {
		fmt.Printf("\t~ Request to Uberduck: %s\n", response.UUID)
		uberduckStatus, err := checkUberduckStatus(voice, response.UUID)

		// if "" == "" {
		if err != nil {
			fmt.Println("Error:", err)
			break
		} else if uberduckStatus.StartedAt == "" {
			fmt.Printf("StartedAt is Blank: %+v | %+v", err, uberduckStatus)
			break
		} else if uberduckStatus.FailedAt != "" {
			fmt.Println("Failed! ", err)
			break
		} else if uberduckStatus.FinishedAt != "" {

			resp, err = http.Get(uberduckStatus.Path)
			if err != nil {
				fmt.Fprintf(
					os.Stderr,
					"Error downloading the WAV file: %v | %s\n",
					err,
					uberduckStatus.Path)
				os.Exit(1)
			}
			defer resp.Body.Close()

			// Save the WAV file to the output file path
			data, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading WAV data: %v\n", err)
				os.Exit(1)
			}

			outputFile := dir + fmt.Sprintf("/../../tmp/voices/%s", voiceFile)
			err = ioutil.WriteFile(outputFile, data, 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error writing the output WAV file: %v\n", err)
				os.Exit(1)
			}

			// 	fmt.Printf("\t~ New Audio File: '%s'\n", outputFile)

			dialogueFile := dir + fmt.Sprintf("/../../tmp/voices/%s.txt", voice)
			fmt.Printf("dialogueFile: %s", dialogueFile)
			err = ioutil.WriteFile(dialogueFile, []byte(fullScript), 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error writing the output TXT file: %v\n", err)
				return
			}

			// TODO: Abstract Finding this
			animate := fmt.Sprintf("/home/begin/code/BeginGPT/animation_scripts/animate.sh")

			fmt.Printf("\n\t~~~~ Starting Animation: \n\t%s\n", animate)

			simulatedCommand := fmt.Sprintf("animation_scripts/animate %s %s %s\n",
				voiceFile,
				dialogueFile,
				animationNamespace,
			)
			fmt.Printf("\n%s\n\n", simulatedCommand)

			cmd := exec.Command("/bin/sh", animate, voiceFile, dialogueFile, animationNamespace)

			var outb, errb bytes.Buffer
			cmd.Stdout = &outb
			cmd.Stderr = &errb
			err = cmd.Run()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Animation out:", outb.String(), "err:", errb.String())

			// We need this to be passed in
			// TODO:
			broadcast <- fmt.Sprintf("start_animation")

			break
		}
		time.Sleep(3 * time.Second)
	}
}

type VoiceText struct {
	Voice string
	Text  string
}

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
