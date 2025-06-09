package audio

import (
	"bytes"
	"os"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

// ToOgg converts audio data of any format to OGG format
func ToOgg(audioData []byte) ([]byte, error) {
	// Create input reader from byte slice
	inputReader := bytes.NewReader(audioData)

	// Create output buffer
	outputBuffer := bytes.NewBuffer(nil)

	// Run ffmpeg conversion to OGG (auto-detect input format)
	err := ffmpeg.Input("pipe:").
		Output("pipe:", ffmpeg.KwArgs{"f": "ogg", "c:a": "libvorbis"}).
		WithInput(inputReader).
		WithOutput(outputBuffer, os.Stderr).
		Run()

	if err != nil {
		return nil, err
	}

	return outputBuffer.Bytes(), nil
}

// ToWav converts audio data of any format to WAV format
func ToWav(audioData []byte) ([]byte, error) {
	// Create input reader from byte slice
	inputReader := bytes.NewReader(audioData)

	// Create output buffer
	outputBuffer := bytes.NewBuffer(nil)

	// Run ffmpeg conversion to WAV (auto-detect input format)
	err := ffmpeg.Input("pipe:").
		Output("pipe:", ffmpeg.KwArgs{"f": "wav"}).
		WithInput(inputReader).
		WithOutput(outputBuffer, os.Stderr).
		Run()

	if err != nil {
		return nil, err
	}

	return outputBuffer.Bytes(), nil
}
