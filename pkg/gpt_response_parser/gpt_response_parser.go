package gpt_response_parser

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"beginbot.com/GoBeginGPT/pkg/skybox"
	"beginbot.com/GoBeginGPT/pkg/uberduck"
)

func getGrandparentDir() (string, error) {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		return "", fmt.Errorf("failed to get caller information")
	}
	parentDir := filepath.Dir(filename)
	grandparentDir := filepath.Dir(parentDir)
	return grandparentDir, nil
}

func SplitDuet(voiceFile string) {
	dir, err := getGrandparentDir()
	// We need to read in the duet
	duetFile := dir + fmt.Sprintf("/tmp/%s", voiceFile)
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

	// lines = strings.Split(strings.Join(lines, " "), verseThreeSplitter)
	verse3 := lines[1]

	// fmt.Println("\n~~ -----------------------")
	// panic(verse3)

	verse1 = strings.ReplaceAll(verse1, chorusLine, "")
	verse1 = strings.ReplaceAll(verse1, verseOneSplitter, "")

	verse2 = strings.ReplaceAll(verse2, chorusLine, "")
	verse2 = strings.ReplaceAll(verse2, verseTwoSplitter, "")

	verse3 = strings.ReplaceAll(verse3, chorusLine, "")
	verse3 = strings.ReplaceAll(verse3, verseThreeSplitter, "")

	// done := make(chan bool, 3)
	done := make(chan bool)
	fmt.Printf("%s", verse1)
	fmt.Printf("---------------------")
	fmt.Printf("%s", verse2)
	fmt.Printf("---------------------")
	fmt.Printf("%s", verse3)
	// GO GO GO!!!!
	// So we need to save different

	streamCharacter := "seal"
	animationNamespace := "verse1"

	fmt.Printf("%s %s", streamCharacter, animationNamespace)
	go uberduck.TextToVoiceAndAnimate(
		streamCharacter,
		"snoop-dogg",
		"verse1.wav",
		"verse1",
		verse1,
	)

	go uberduck.TextToVoiceAndAnimate(
		streamCharacter,
		"2pac",
		"verse2.wav",
		"verse2",
		verse2,
	)

	go uberduck.TextToVoiceAndAnimate(
		streamCharacter,
		"snoop-dogg",
		"verse3.wav",
		"verse3",
		verse3,
	)

	// When does the browser refresh cache

	<-done
	// How do we make sure this waits for all
	// I could pass in a chan with 3 message
	// Now I need to process each of these
}

func SplitScript() string {
	dir, _ := getGrandparentDir()
	fileName := dir + "/tmp/chatgpt_response.txt"

	lines, err := readFile(fileName)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return ""
	}

	script, stageDirection := splitAndSort(lines)

	if len(stageDirection) > 1 {
		prompt := stageDirection[0]
		fmt.Println("prompt: %s", prompt)
		go skybox.GenerateSkybox(prompt)
	}

	// So now we recombine:

	fullScript := strings.Join(script, " ")
	return fullScript
	// for _, line := range script {
	// 	fmt.Println(line)
	// }
}

func readFile(fileName string) ([]string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func splitAndSort(lines []string) ([]string, []string) {
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
