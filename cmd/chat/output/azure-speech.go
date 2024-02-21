package output

import (
	"fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/go-resty/resty/v2"
	"os"
	"time"
)

type AzureSpeechOutput struct {
	client       *resty.Client
	speaker      string
	outputFormat string
}

func NewAzureSpeechOutput(speaker string) Output {
	speechKey := os.Getenv("SPEECH_KEY")
	speechUrl := os.Getenv("SPEECH_URL")

	client := resty.New()

	client.SetDebug(false)
	client.SetHeader("Ocp-Apim-Subscription-Key", speechKey)
	client.SetBaseURL(speechUrl)

	return &AzureSpeechOutput{
		client:       client,
		speaker:      speaker,
		outputFormat: "audio-24khz-48kbitrate-mono-mp3",
	}
}

func (a *AzureSpeechOutput) Run(input string) error {
	requestBody := fmt.Sprintf(
		`<speak version='1.0' xml:lang='zh-CN'><voice xml:lang='zh-CN' xml:gender='Male'
    name='%s'>
       %s
</voice></speak>`, a.speaker, input)

	response, err := a.client.R().
		SetHeader("Content-Type", "application/ssml+xml").
		SetHeader("X-Microsoft-OutputFormat", a.outputFormat).
		SetBody(requestBody).
		Post("cognitiveservices/v1")

	if err != nil {
		return err
	}

	if !response.IsSuccess() {
		return fmt.Errorf("error: %s", response.String())
	}

	if len(response.Body()) == 0 {
		return fmt.Errorf("no audio returned")
	}

	err = a.playAudio(response)
	if err != nil {
		return err
	}

	return nil
}

func (a *AzureSpeechOutput) playAudio(response *resty.Response) error {
	// save the response to a file
	err := os.WriteFile("output.mp3", response.Body(), 0644)
	if err != nil {
		return err
	}

	f, err := os.Open("output.mp3")
	if err != nil {
		return err
	}

	streamer, formap, err := mp3.Decode(f)
	if err != nil {
		return err
	}

	defer streamer.Close()

	speaker.Init(formap.SampleRate, formap.SampleRate.N(time.Second/10))

	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))

	<-done
	return nil
}
