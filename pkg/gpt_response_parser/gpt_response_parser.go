package gpt_response_parser

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"beginbot.com/GoBeginGPT/pkg/skybox"
	"beginbot.com/GoBeginGPT/pkg/uberduck"
	"beginbot.com/GoBeginGPT/pkg/utils"
)

var outerDir, _ = utils.GetGreatGreatGrandparentDir()
var dir, _ = utils.GetGreatGrandparentDir()

var combineScript = "/home/begin/code/BeginGPT/animate_scripts/combine_videos.sh"
var ffmpegFileListPath = outerDir + "/animate_scripts/videos.txt"
var chatgptResponseFilepath = outerDir + "/tmp/duet.txt"

// var chatgptResponseFilepath = dir + "/tmp/chatgpt_response.txt"

func SplitDuet(broadcast chan string, voiceFile string) {
	duetFile := outerDir + fmt.Sprintf("/tmp/current/%s", voiceFile)

	duet, err := ioutil.ReadFile(duetFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading Duet file: %v\n", err)
		os.Exit(1)
	}

	verseOneSplitter := "(Verse 1: Snoop Dogg)"
	verseTwoSplitter := "(Verse 2: Tupac)"
	verseThreeSplitter := "(Verse 3: Snoop Dogg)"
	chorusLine := "Chorus:"

	lines := strings.Split(string(duet), verseTwoSplitter)
	verse1 := lines[0]

	lines = strings.Split(lines[1], verseTwoSplitter)
	verse2 := lines[0]
	lines = strings.Split(verse2, verseThreeSplitter)
	verse2 = lines[0]

	verse3 := lines[1]

	verse1 = strings.ReplaceAll(verse1, chorusLine, "")
	verse1 = strings.ReplaceAll(verse1, verseOneSplitter, "")

	verse2 = strings.ReplaceAll(verse2, chorusLine, "")
	verse2 = strings.ReplaceAll(verse2, verseTwoSplitter, "")

	verse3 = strings.ReplaceAll(verse3, chorusLine, "")
	verse3 = strings.ReplaceAll(verse3, verseThreeSplitter, "")

	// fmt.Printf("%s", verse1)
	// fmt.Printf("---------------------")
	// fmt.Printf("%s", verse2)
	// fmt.Printf("---------------------")
	// fmt.Printf("%s", verse3)

	streamCharacter := "seal"
	animationNamespace := "duet"

	type Verse struct {
		Name      string
		Content   string
		Character string
		Voice     string
	}

	verses := []Verse{
		{
			Content:   verse1,
			Voice:     "snoop-dogg",
			Character: streamCharacter,
		},
		{
			Content:   verse2,
			Voice:     "2pac",
			Character: streamCharacter,
		},
		{
			Content:   verse3,
			Voice:     "snoop-dogg",
			Character: streamCharacter,
		},
	}

	var Files = []string{}
	var wg sync.WaitGroup

	animationFolder, err := filepath.Abs(outerDir + "/" + animationNamespace)
	if err != nil {
		fmt.Printf("Error Finding Absolute animation Folder: %+v", err)
		return
	}

	err = os.MkdirAll(animationFolder, os.ModePerm)

	if err != nil {
		fmt.Printf("Erroring Mkdir: %+v", err)
	}

	for i, verse := range verses {
		wg.Add(1)
		voiceFile := fmt.Sprintf("%s_%d", verse.Voice, i)

		go uberduck.TextToVoiceAndAnimate(
			broadcast,
			verse.Character,
			verse.Voice,
			voiceFile,
			animationNamespace,
			verse.Content,
			&wg,
		)

		media, _ := filepath.Abs(dir + fmt.Sprintf("/static/media/%s.mp4", voiceFile))
		Files = append(Files, media)
	}

	wg.Wait()
	// ===========================================

	// This can't be called until all 3 are done
	tmpl, err := template.New("ffmpeg File visit").Parse("{{range $val := .}} file '{{$val}}'\n{{end}}")
	if err != nil {
		panic(err)
	}

	var ffmpegFileList bytes.Buffer
	err = tmpl.Execute(&ffmpegFileList, Files)
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(ffmpegFileListPath, []byte(ffmpegFileList.String()), 0644)

	cmd := exec.Command("/bin/sh", combineScript)
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func SplitScript() string {
	lines, err := utils.ReadFile(chatgptResponseFilepath)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return ""
	}

	script, stageDirection := splitAndSortParenthesis(lines)

	if len(stageDirection) > 1 {
		prompt := stageDirection[0]
		fmt.Printf("prompt: %s\n", prompt)
		go skybox.GenerateSkybox(prompt)
	}

	fullScript := strings.Join(script, " ")
	return fullScript
}

func splitAndSortParenthesis(lines []string) ([]string, []string) {
	var parenthLines []string
	var nonParenthLines []string

	parenthRegex := regexp.MustCompile(`^\([^()]*\)$`)

	for _, line := range lines {
		if parenthRegex.MatchString(line) {
			parenthLines = append(parenthLines, line)
		} else {
			nonParenthLines = append(nonParenthLines, line)
		}
	}
	return nonParenthLines, parenthLines
}
