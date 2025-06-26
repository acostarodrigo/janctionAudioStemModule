package keeper

import (
	"context"
	"errors"
	"log"

	"cosmossdk.io/collections"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/janction/audioStem"
)

var _ audioStem.QueryServer = queryServer{}

// NewQueryServerImpl returns an implementation of the module QueryServer.
func NewQueryServerImpl(k Keeper) audioStem.QueryServer {
	return queryServer{k}
}

type queryServer struct {
	k Keeper
}

// GetGame defines the handler for the Query/GetGame RPC method.
func (qs queryServer) GetAudioStemTask(ctx context.Context, req *audioStem.QueryGetAudioStemTaskRequest) (*audioStem.QueryGetAudioStemTaskResponse, error) {
	audioStemTask, err := qs.k.AudioStemTasks.Get(ctx, req.Index)
	if err == nil {
		return &audioStem.QueryGetAudioStemTaskResponse{AudioStemTask: &audioStemTask}, nil
	}
	if errors.Is(err, collections.ErrNotFound) {
		return &audioStem.QueryGetAudioStemTaskResponse{AudioStemTask: nil}, nil
	}

	return nil, status.Error(codes.Internal, err.Error())
}

func (qs queryServer) GetAudioStemLogs(ctx context.Context, req *audioStem.QueryGetAudioStemLogsRequest) (*audioStem.QueryGetAudioStemLogsResponse, error) {
	// access database
	var logs []*audioStem.AudioStemLogs_AudioStemLog
	result := qs.k.DB.ReadLogs(req.ThreadId)
	if len(result) == 0 {
		return nil, nil
	}
	for _, val := range result {
		logEntry := audioStem.AudioStemLogs_AudioStemLog{Log: val.Log, Timestamp: val.Timestamp, Severity: audioStem.AudioStemLogs_AudioStemLog_SEVERITY(val.Severity)}
		logs = append(logs, &logEntry)
	}

	return &audioStem.QueryGetAudioStemLogsResponse{AudioStemLogs: &audioStem.AudioStemLogs{ThreadId: req.ThreadId, Logs: logs}}, nil
}

func (qs queryServer) GetPendingAudioStemTasks(ctx context.Context, req *audioStem.QueryGetPendingAudioStemTaskRequest) (*audioStem.QueryGetPendingAudioStemTaskResponse, error) {
	ti, err := qs.k.AudioStemTaskInfo.Get(ctx)

	if err != nil {
		return nil, err
	}
	nextId := ti.NextId

	var result []*audioStem.AudioStemTask
	for i := 0; i < int(nextId); i++ {
		task, err := qs.k.AudioStemTasks.Get(ctx, string(i))
		if err != nil {
			log.Fatalf("unable to retrieve task with id %v. Error: %v", string(i), err.Error())
			continue
		}

		if !task.Completed {
			result = append(result, &task)
		}
	}
	return &audioStem.QueryGetPendingAudioStemTaskResponse{AudioStemTasks: result}, nil
}

func (qs queryServer) GetWorker(ctx context.Context, req *audioStem.QueryGetWorkerRequest) (*audioStem.QueryGetWorkerResponse, error) {
	worker, err := qs.k.Workers.Get(ctx, req.Worker)
	if err != nil {
		return nil, err
	}

	return &audioStem.QueryGetWorkerResponse{Worker: &worker}, nil
}
