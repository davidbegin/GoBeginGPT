package utils

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

func GetGreatGreatGrandparentDir() (string, error) {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		return "", fmt.Errorf("failed to get caller information")
	}
	parentDir := filepath.Dir(filename)
	grandparentDir := filepath.Dir(parentDir)
	greatGrandparentDir := filepath.Dir(grandparentDir)
	ggpd := filepath.Dir(greatGrandparentDir)
	return ggpd, nil
}

func GetGreatGrandparentDir() (string, error) {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		return "", fmt.Errorf("failed to get caller information")
	}
	parentDir := filepath.Dir(filename)
	grandparentDir := filepath.Dir(parentDir)
	greatGrandparentDir := filepath.Dir(grandparentDir)
	return greatGrandparentDir, nil
}

func GetGrandparentDir() (string, error) {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		return "", fmt.Errorf("failed to get caller information")
	}
	parentDir := filepath.Dir(filename)
	grandparentDir := filepath.Dir(parentDir)
	return grandparentDir, nil
}

// Do I need this file in here????
func ReadFile(fileName string) ([]string, error) {
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

// TODO: Why isn't this working?
func Talk(soundFile string) string {
	fmt.Printf("Talk Time!\n")
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

func MdToHTML(md []byte) []byte {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return markdown.Render(doc, renderer)
}
