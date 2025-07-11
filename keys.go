package audioStem

import "cosmossdk.io/collections"

const ModuleName = "audioStem"

var (
	ParamsKey                = collections.NewPrefix("audioStemParams")
	AudioStemTaskKey         = collections.NewPrefix("audioStemTaskList/value/")
	WorkerKey                = collections.NewPrefix("audioStemWorker")
	TaskInfoKey              = collections.NewPrefix(0)
	PendingAudioStemTasksKey = collections.NewPrefix(1)
)
