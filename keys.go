package audioStem

import "cosmossdk.io/collections"

const ModuleName = "audioStem"

var (
	ParamsKey                = collections.NewPrefix("Params")
	AudioStemTaskKey         = collections.NewPrefix("audioStemTaskList/value/")
	WorkerKey                = collections.NewPrefix("Worker")
	TaskInfoKey              = collections.NewPrefix(0)
	PendingAudioStemTasksKey = collections.NewPrefix(1)
)
