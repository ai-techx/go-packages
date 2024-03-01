package plugins

import (
	"fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/go-resty/resty/v2"
	"github.com/meta-metopia/go-packages/pkg/ai/gpt/dto"
	"github.com/meta-metopia/go-packages/pkg/ai/gpt/plugin"
	"os"
	"time"
)

type AzurePlugin struct {
	plugin.Client
	client       *resty.Client
	speaker      string
	outputFormat string
}

func (c *AzurePlugin) Name() string {
	return "AzurePlugin"
}

func (c *AzurePlugin) Description() string {
	return "AzurePlugin"
}

func (c *AzurePlugin) ConvertOutput(response dto.Message) (*plugin.ConvertedResponse, error) {
	if response.Content == "" {
		return nil, nil
	}

	requestBody := fmt.Sprintf(
		`<speak version='1.0' xml:lang='zh-CN'><voice xml:lang='zh-CN' xml:gender='Male'
    name='%s'>
       %s
</voice></speak>`, c.speaker, response.Content)

	outputResponse, err := c.client.R().
		SetHeader("Content-Type", "application/ssml+xml").
		SetHeader("X-Microsoft-OutputFormat", c.outputFormat).
		SetBody(requestBody).Post("cognitiveservices/v1")

	if err != nil {
		return nil, err
	}

	if !outputResponse.IsSuccess() {
		return nil, fmt.Errorf("error: %s", outputResponse.String())
	}

	if len(outputResponse.Body()) == 0 {
		return nil, fmt.Errorf("no audio returned")
	}

	err = c.playAudio(outputResponse)
	if err != nil {
		return nil, err
	}

	return &plugin.ConvertedResponse{
		Action: plugin.AddToOutputAfter,
		Message: &dto.Message{
			Content: "Mock Output",
		},
	}, nil
}

func (a *AzurePlugin) playAudio(response *resty.Response) error {
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

func NewAzurePlugin(speaker string) plugin.Interface {
	speechKey := os.Getenv("SPEECH_KEY")
	speechUrl := os.Getenv("SPEECH_URL")

	client := resty.New()

	client.SetDebug(false)
	client.SetHeader("Ocp-Apim-Subscription-Key", speechKey)
	client.SetBaseURL(speechUrl)

	return &AzurePlugin{
		client:       client,
		speaker:      speaker,
		outputFormat: "audio-24khz-48kbitrate-mono-mp3",
	}
}
