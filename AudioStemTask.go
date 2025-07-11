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

func (t *AudioStemTask) GenerateThreads(taskId string, cid string) (res []*AudioStemThread) {
	for i := range t.AmountFiles {
		thread := AudioStemThread{ThreadId: t.TaskId + strconv.FormatInt(int64(i), 10), TaskId: taskId, Instrument: t.Instrument, Mp3: t.Mp3, Cid: cid}
		res = append(res, &thread)
	}
	return res
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
