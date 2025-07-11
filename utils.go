package audioStem

import (
	"crypto/sha256"
	"encoding/hex"
	fmt "fmt"
	io "io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-audio/wav"
	"github.com/hajimehoshi/go-mp3"

	"github.com/janction/audioStem/audioStemLogger"
)

// Transforms a slice with format [key]=[value] to a map
func TransformSliceToMap(input []string) (map[string]string, error) {
	result := make(map[string]string)

	for _, item := range input {
		parts := strings.SplitN(item, "=", 2) // Split into 2 parts: filename and hash
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid format: %s", item)
		}
		filename := parts[0]
		hash := parts[1]
		result[filename] = hash
	}

	return result, nil
}

// MapToKeyValueFormat converts a map[string]string to a "key=value,key=value" format
func MapToKeyValueFormat(inputMap map[string]string) []string {
	var parts []string

	// Iterate through the map and build the key=value pairs
	for key, value := range inputMap {
		parts = append(parts, fmt.Sprintf("%s=%s", key, value))
	}

	// Join the key=value pairs with commas
	return parts
}

// Executes a cli command with their arguments
func ExecuteCli(args []string) error {
	executableName := "janctiond"
	args = append(args, "--gas")
	args = append(args, "auto")
	args = append(args, "--gas-adjustment")
	args = append(args, "1.3")
	cmd := exec.Command(executableName, args...)
	audioStemLogger.Logger.Debug("Executing %s", cmd.String())

	_, err := cmd.Output()

	if err != nil {
		audioStemLogger.Logger.Error("Error Executing CLI %s: %s", cmd.String(), err.Error())
		return err
	}

	return nil
}

func FromCliToFrames(entries []string) map[string]AudioStemThread_Stem {
	result := make(map[string]AudioStemThread_Stem)

	for _, entry := range entries {
		parts := strings.Split(entry, "=")
		if len(parts) != 2 {
			fmt.Println("Invalid entry:", entry)
			continue
		}

		filename := parts[0]
		cidAndHash := strings.Split(parts[1], ":")
		if len(cidAndHash) != 2 {
			fmt.Println("Invalid CID:Hash format:", parts[1])
			continue
		}
		frame := AudioStemThread_Stem{Filename: filename, Cid: cidAndHash[0], Hash: cidAndHash[1]}
		result[filename] = frame
	}

	return result
}

func FromFramesToCli(frames map[string]AudioStemThread_Stem) []string {
	var result []string

	for filename, frame := range frames {
		entry := fmt.Sprintf("%s=%s:%s", filename, frame.Cid, frame.Hash)
		result = append(result, entry)
	}

	return result
}

// CalculateFileHash calculates the SHA-256 hash of a given file.
func CalculateFileHash(filePath string) (string, error) {
	hash, err := calculateAudioSampleHash(filePath)
	if err != nil {
		return "", err
	}
	return hash, nil
}

func calculateAudioSampleHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hasher := sha256.New()

	// Try WAV decoding first
	file.Seek(0, io.SeekStart)
	wavDecoder := wav.NewDecoder(file)

	buf, err := wavDecoder.FullPCMBuffer()
	if err == nil && buf != nil && len(buf.Data) > 0 {
		for _, sample := range buf.Data {
			hasher.Write([]byte{byte(sample >> 8), byte(sample)})
		}
		return hex.EncodeToString(hasher.Sum(nil)), nil
	}

	// Reset and fallback to MP3 only if WAV decoding failed
	file.Seek(0, io.SeekStart)
	mp3Decoder, err := mp3.NewDecoder(file)
	if err != nil {
		return "", fmt.Errorf("failed to decode WAV or MP3: %w", err)
	}

	bufBytes := make([]byte, 4096)
	for {
		n, err := mp3Decoder.Read(bufBytes)
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("error reading MP3: %w", err)
		}
		hasher.Write(bufBytes[:n])
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// GenerateDirectoryFileHashes walks through a directory and computes SHA-256 hashes for all files.
func GenerateDirectoryFileHashes(dirPath string) (map[string]string, error) {
	hashes := make(map[string]string)

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Compute file hash
		hash, err := CalculateFileHash(path)
		if err != nil {
			return err
		}

		// Store hash with filename (relative path)
		relPath, _ := filepath.Rel(dirPath, path)
		hashes[relPath] = hash

		return nil
	})

	if err != nil {
		return nil, err
	}

	return hashes, nil
}
