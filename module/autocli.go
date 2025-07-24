package module

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	audioStemv1 "github.com/janction/audioStem/api/v1"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: audioStemv1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "GetAudioStemTask",
					Use:       "get-audio-stem-task index",
					Short:     "Get the current value of the Audio Stem task at index",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "index"},
					},
				},
				{
					RpcMethod: "GetPendingAudioStemTasks",
					Use:       "get-pending-audio-stem-tasks",
					Short:     "Gets the pending audio stem tasks",
				},
				{
					RpcMethod: "GetWorker",
					Use:       "get-worker worker",
					Short:     "Gets a single worker",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "worker"},
					},
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: audioStemv1.Msg_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "CreateAudioStemTask",
					Use:       "create-audio-stem-task [cid] [amountFiles] [instrument] [mp3] [reward]",
					Short:     "Creates a new audio stem task",
					Long:      "", // TODO Add long
					Example:   "", // TODO add exampe
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "cid"},
						{ProtoField: "amount_files"},
						{ProtoField: "instrument"},
						{ProtoField: "mp3"},
						{ProtoField: "reward"},
					},
				},
				{
					RpcMethod: "AddWorker",
					Use:       "add-worker [public_ip] [ipfs_id] [stake]--from [workerAddress]",
					Short:     "Registers a new worker that will perform audio stem tasks",
					Long:      "", // TODO Add long
					Example:   "", // TODO add exampe
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "public_ip"},
						{ProtoField: "ipfs_id"},
						{ProtoField: "stake"},
					},
				},
				{
					RpcMethod: "SubscribeWorkerToTask",
					Use:       "subscribe-worker-to-task [address] [taskId] [threadId] --from [workerAddress]",
					Short:     "Subscribes an existing enabled worker to perform work in the specified task",
					Long:      "", // TODO Add long
					Example:   "", // TODO add exampe
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "address"},
						{ProtoField: "taskId"},
						{ProtoField: "threadId"},
					},
				},
				{
					RpcMethod: "ProposeSolution",
					Use:       "propose-solution [taskId] [threadId] [publicKey] [signatures] --from [workerAddress]",
					Short:     "Proposes a solution to a thread.",
					Long:      "", // TODO Add long
					Example:   "", // TODO add exampe
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "taskId"},
						{ProtoField: "threadId"},
						{ProtoField: "public_key"},
						{ProtoField: "signatures", Varargs: true},
					},
				},
				{
					RpcMethod: "SubmitSolution",
					Use:       "submit-solution [taskId] [threadId] [cid] [average_render_seconds] --from [workerAddress]",
					Short:     "Submits the cid of the directory with all the uploaded frames.",
					Long:      "", // TODO Add long
					Example:   "", // TODO add exampe
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "taskId"},
						{ProtoField: "threadId"},
						{ProtoField: "dir", Varargs: false},
						{ProtoField: "average_stem_seconds"},
					},
				},
				{
					RpcMethod: "SubmitValidation",
					Use:       "submit-validation [taskId] [threadId] [publicKey] [signatures] --from [workerAddress]",
					Short:     "Submit a validation to a proposed solution",
					Long:      "", // TODO Add long
					Example:   "", // TODO add exampe
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "taskId"},
						{ProtoField: "threadId"},
						{ProtoField: "public_key"},
						{ProtoField: "signatures", Varargs: true},
					},
				},
				{
					RpcMethod: "RevealSolution",
					Use:       "reveal-solution [taskId] [threadId] [solution] --from [workerAddress]",
					Short:     "Reveals the CiDs of the solution",
					Long:      "", // TODO Add long
					Example:   "", // TODO add exampe
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "taskId"},
						{ProtoField: "threadId"},
						{ProtoField: "stems", Varargs: true},
					},
				},
			},
		},
	}
}
