package audioStem

import (
	"context"
	"strconv"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/janction/audioStem/db"
)

func (AudioStemTask) Validate() error {
	return nil
}

func GetEmptyAudioStemTaskList() []IndexedAudioStemTask {
	return []IndexedAudioStemTask{}
}

func (t *AudioStemTask) GenerateThreads(taskId string) (res []*AudioStemThread) {
	// Split frames among the threads
	frameRanges := splitFrames(int(t.StartFrame), int(t.EndFrame), int(t.ThreadAmount))

	// Print the result

	for i, r := range frameRanges {
		thread := AudioStemThread{ThreadId: t.TaskId + strconv.FormatInt(int64(i), 10), StartFrame: int64(r.StartFrame), EndFrame: int64(r.EndFrame), TaskId: taskId}
		res = append(res, &thread)
	}

	return res
}

// SplitFrames divides the total frames into chunks based on the number of threads
// FrameRange represents the range of frames assigned to a thread
type frameRange struct {
	StartFrame int
	EndFrame   int
}

func splitFrames(startFrame, endFrame, threads int) []frameRange {
	totalFrames := endFrame - startFrame + 1 // Total number of frames
	framesPerThread := totalFrames / threads // Base number of frames per thread
	remainder := totalFrames % threads       // Remaining frames to distribute

	result := make([]frameRange, threads)

	currentStart := startFrame
	for i := 0; i < threads; i++ {
		extra := 0
		if remainder > 0 {
			extra = 1
			remainder--
		}

		end := currentStart + framesPerThread + extra - 1
		result[i] = frameRange{StartFrame: currentStart, EndFrame: end}
		currentStart = end + 1
	}

	return result
}

func (t AudioStemTask) SubscribeWorkerToTask(ctx context.Context, workerAddress, taskId, threadId string, db db.Database) error {
	// TODO call cmd with message subscribeWorkerToTask
	args := []string{
		"tx", "audioStem", "subscribe-worker-to-task",
		workerAddress, taskId, threadId, "--yes", "--from", workerAddress,
	}
	err := ExecuteCli(args)
	if err != nil {
		db.UpdateTask(taskId, threadId, false)
		return err
	}
	return nil
}

func (t *AudioStemTask) GetWinnerReward() types.Coin {
	amountThreads := len(t.Threads)
	return types.NewCoin(t.Reward.Denom, t.Reward.Amount.QuoRaw(2).QuoRaw(int64(amountThreads)))
}

func (t *AudioStemTask) GetValidatorsReward() types.Coin {
	amountThreads := len(t.Threads)
	return types.NewCoin(t.Reward.Denom, t.Reward.Amount.QuoRaw(2).QuoRaw(int64(amountThreads)))
}
