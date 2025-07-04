package vm

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/janction/audioStem/audioStemLogger"
	"github.com/janction/audioStem/db"
)

func IsContainerRunning(ctx context.Context, threadId string) bool {
	name := fmt.Sprintf("janctionstem%s", threadId)

	// Command to check for running containers
	cmd := exec.CommandContext(ctx, "docker", "ps", "--filter", fmt.Sprintf("name=%s", name), "--format", "{{.Names}}")

	output, err := cmd.Output()
	if err != nil {
		audioStemLogger.Logger.Error("Error executing Docker command: %v\n", err)
		return false
	}

	// Trim output and compare with container name
	containerName := strings.TrimSpace(string(output))
	return containerName == name
}

func StemAudio(ctx context.Context, id string, filename string, instrument string, mp3 bool, path string, db *db.DB) error {
	n := "janctionstem" + id

	started := time.Now().Unix()
	db.AddLogEntry(id, fmt.Sprintf("Started stem file %v...", filename), started, 0)

	// Check if the container exists using `docker ps -a`
	checkCmd := exec.CommandContext(ctx, "docker", "ps", "-a", "--filter", fmt.Sprintf("name=%s", n), "--format", "{{.Names}}")
	output, err := checkCmd.Output()
	if err != nil {
		db.AddLogEntry(id, "Error trying to verify if container already exists.", started, 2)
		fail := fmt.Errorf("failed to check container existence: %w", err)
		audioStemLogger.Logger.Error(fail.Error())
		return fail
	}

	// If the container already exists, exit the function
	if string(output) != "" {
		audioStemLogger.Logger.Debug("Container already exists.")
		return nil
	}

	// Construct the bind path and command
	bindPathInput := fmt.Sprintf("%s:/data/input", path)
	bindPathOutput := fmt.Sprintf("%s:/data/output", path)
	audioPath := fmt.Sprintf("/data/input/%s", filename)

	var stemArgs []string
	stemArgs = append(stemArgs, "-n")
	stemArgs = append(stemArgs, "htdemucs")
	stemArgs = append(stemArgs, "--out")
	stemArgs = append(stemArgs, "/data/output")
	stemArgs = append(stemArgs, "--shifts")
	stemArgs = append(stemArgs, "1")

	stemArgs = append(stemArgs, "--overlap")
	stemArgs = append(stemArgs, "0.25")
	stemArgs = append(stemArgs, "-j")
	stemArgs = append(stemArgs, "1")
	// stemArgs = append(stemArgs, fmt.Sprintf("\"%s\"", audioPath))
	stemArgs = append(stemArgs, audioPath) // no quotes

	var dockerArgs []string
	dockerArgs = append(dockerArgs, "run")

	dockerArgs = append(dockerArgs, "--rm")
	// dockerArgs = append(dockerArgs, "-it")
	dockerArgs = append(dockerArgs, "--name")
	dockerArgs = append(dockerArgs, n)
	dockerArgs = append(dockerArgs, "-v")
	dockerArgs = append(dockerArgs, bindPathInput)
	dockerArgs = append(dockerArgs, "-v")
	dockerArgs = append(dockerArgs, bindPathOutput)

	dockerArgs = append(dockerArgs, "janction/audio-stem:latest")
	dockerArgs = append(dockerArgs, stemArgs...)

	// Create and start the container
	runCmd := exec.CommandContext(ctx, "docker", dockerArgs...)
	audioStemLogger.Logger.Info("Starting docker: %s", runCmd.String())
	output, err = runCmd.CombinedOutput()
	if err != nil {
		db.AddLogEntry(id, fmt.Sprintf("Error in creating the container. %s", err.Error()), started, 1)
		audioStemLogger.Logger.Error("failed to run container: %s\nOutput:\n%s", err.Error(), string(output))
		return fmt.Errorf("failed to create and start container: %w", err)
	}

	// Wait for the container to finish
	waitCmd := exec.CommandContext(ctx, "docker", "wait", n)
	err = waitCmd.Run()
	if err != nil {
		audioStemLogger.Logger.Error("failed to wait for container: %s", err.Error())
		return fmt.Errorf("failed to wait for container: %w", err)
	}

	// Retrieve and print logs
	logsCmd := exec.CommandContext(ctx, "docker", "logs", n)
	logsOutput, err := logsCmd.Output()
	if err != nil {
		audioStemLogger.Logger.Error("failed to retrieve container logs: %s", err.Error())
		return fmt.Errorf("failed to retrieve container logs: %w", err)
	}
	audioStemLogger.Logger.Info("Container logs:")
	audioStemLogger.Logger.Info(string(logsOutput))

	RemoveContainer(ctx, n)

	// Verify the frame exists and log
	// frameFile := FormatFrameFilename(int(frameNumber))
	// framePath := filepath.Join(path, "htdemucs", frameFile)
	// finish := time.Now().Unix()
	// difference := time.Unix(finish, 0).Sub(time.Unix(started, 0))
	// if _, err := os.Stat(framePath); errors.Is(err, os.ErrNotExist) {
	// 	db.AddLogEntry(id, fmt.Sprintf("Error while rendering frame %v. %s file is not there", frameNumber, framePath), started, 2)
	// 	renderVideoFrame(ctx, cid, frameNumber, id, path, db)
	// } else {
	// 	// we capture the duration of the rendering
	// 	duration := int(difference.Seconds())
	// 	// we add the log
	// 	db.AddLogEntry(id, fmt.Sprintf("Successfully rendered frame %v in %v seconds.", frameNumber, duration), finish, 1)
	// 	// and record the duration for the frame
	// 	audioStemLogger.Logger.Info("Recorded duration for frame %v: %v seconds", int(frameNumber), duration)
	// 	db.AddRenderDuration(id, int(frameNumber), duration)
	// }
	return nil
}

func RemoveContainer(ctx context.Context, name string) error {
	// Remove the container after completion
	rmCmd := exec.CommandContext(ctx, "docker", "rm", name)
	err := rmCmd.Run()
	if err != nil {
		audioStemLogger.Logger.Error(err.Error())
	}
	return err
}

// CountFilesInDirectory counts the number of files in a given directory
func CountFilesInDirectory(directoryPath string) int {
	// Read the directory contents
	files, err := os.ReadDir(directoryPath)
	if err != nil {
		audioStemLogger.Logger.Error(err.Error())
		return 0
	}

	// Count only files (not subdirectories)
	fileCount := 0
	for _, file := range files {
		if !file.IsDir() {
			fileCount++
		}
	}
	return fileCount
}

// FormatFrameFilename returns the correct filename for a given frame number.
func FormatFrameFilename(frameNumber int) string {
	return fmt.Sprintf("frame_%06d.png", frameNumber)
}

func isARM64() bool {
	audioStemLogger.Logger.Debug("isARM64: %s", runtime.GOARCH)
	return runtime.GOARCH == "arm64"
}

func IsContainerExited(threadId string) (bool, error) {
	containerName := "myBlender" + threadId
	cmd := exec.Command(
		"docker", "ps", "-a",
		"--filter", "name="+containerName,
		"--filter", "status=exited",
		"--format", "{{.Names}}",
	)

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return false, err
	}

	// Trim and check if the container name appears in output
	result := strings.TrimSpace(out.String())
	return result == containerName, nil
}
