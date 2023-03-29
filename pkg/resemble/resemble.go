package resemble

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/resemble-ai/resemble-go/v2"
	"github.com/resemble-ai/resemble-go/v2/request"
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

func generateAuthHeader(apiKey, apiSecret string) string {
	auth := apiKey + ":" + apiSecret
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))
	return "Basic " + encodedAuth
}

func TextToVoiceWithResemble(content []byte) {
	var apiKey string = os.Getenv("RESEMBLE_API_KEY")
	client := resemble.NewClient(apiKey)

	// voiceUUId := "48d7ed16"
	// voiceUUId := "25c7823f"
	voiceUUId := "7f1da153"

	dir, _ := getGrandparentDir()
	recordingFile := dir + "/tmp/uberduck.wav"

	r, err := client.Recording.Create(voiceUUId, recordingFile, request.Payload{
		"name":      "test recording",
		"text":      "transcription",
		"is_active": true,
		"emotion":   "neutral",
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v \n", r)
}
