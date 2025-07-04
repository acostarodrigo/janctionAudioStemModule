package audioStem

import "cosmossdk.io/collections"

const ModuleName = "audioStem"

var (
	ParamsKey                = collections.NewPrefix("AudioStemParams")
	AudioStemTaskKey         = collections.NewPrefix("AudioStemTaskList/value/")
	WorkerKey                = collections.NewPrefix("AudioStemWorker")
	TaskInfoKey              = collections.NewPrefix(0)
	PendingAudioStemTasksKey = collections.NewPrefix(1)
)
