package keeper

import (
	"context"
	"slices"
	"strconv"
	"strings"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ipfs/go-cid"

	"github.com/janction/audioStem"
	"github.com/janction/audioStem/audioStemLogger"
	audioStemCrypto "github.com/janction/audioStem/crypto"
	"github.com/janction/audioStem/ipfs"
)

type msgServer struct {
	k Keeper
}

var _ audioStem.MsgServer = msgServer{}

// NewMsgServerImpl returns an implementation of the module MsgServer interface.
func NewMsgServerImpl(keeper Keeper) audioStem.MsgServer {
	return &msgServer{k: keeper}
}

// CreateGame defines the handler for the MsgCreateAudioStemTask message.
func (ms msgServer) CreateAudioStemTask(ctx context.Context, msg *audioStem.MsgCreateAudioStemTask) (*audioStem.MsgCreateAudioStemTaskResponse, error) {
	audioStemLogger.Logger.Info("CreateAudioStemTask -  creator: %s, cid: %s, startFrame: %v, endFrame: %v, threads: %v, reward: %s", msg.Creator, msg.Cid, msg.StartFrame, msg.EndFrame, msg.Threads, msg.Reward)

	// TODO had validations about the parameters
	taskInfo, err := ms.k.AudioStemTaskInfo.Get(ctx)
	if err != nil {
		audioStemLogger.Logger.Error("Getting task: %s", err.Error())
		return nil, err
	}

	// TODO reward must be valid

	// we make sure the provided CID is valid
	_, err = cid.Decode(msg.Cid)
	if err != nil {
		audioStemLogger.Logger.Error("provided cid is invalid: (%s)", msg.Cid)
		return nil, sdkerrors.ErrAppConfig.Wrapf(audioStem.ErrInvalidAudioStemTask.Error(), "cid %s is invalid", msg.Cid)
	}

	var nextId = taskInfo.NextId
	// we get the taskId in string
	taskId := strconv.FormatInt(nextId, 10)

	// and increase the task id counter for next task
	nextId++
	ms.k.AudioStemTaskInfo.Set(ctx, audioStem.AudioStemTaskInfo{NextId: nextId})

	videoTask := audioStem.AudioStemTask{TaskId: taskId, Requester: msg.Creator, Cid: msg.Cid, StartFrame: msg.StartFrame, EndFrame: msg.EndFrame, Completed: false, ThreadAmount: msg.Threads, Reward: msg.Reward}
	threads := videoTask.GenerateThreads(taskId)
	videoTask.Threads = threads

	// the module will keep the reward to be distributed later
	// TODO Add validations for msg.Creator
	addr, _ := types.AccAddressFromBech32(msg.Creator)
	ms.k.BankKeeper.SendCoinsFromAccountToModule(ctx, addr, audioStem.ModuleName, types.NewCoins(*msg.Reward))

	// we create the task
	if err := ms.k.AudioStemTasks.Set(ctx, taskId, videoTask); err != nil {
		return nil, err
	}
	return &audioStem.MsgCreateAudioStemTaskResponse{TaskId: taskId}, nil
}

func (ms msgServer) AddWorker(ctx context.Context, msg *audioStem.MsgAddWorker) (*audioStem.MsgAddWorkerResponse, error) {
	audioStemLogger.Logger.Info("AddWorker - creator: %s, ipfsId: %s, publicIp: %s, stake: %s", msg.Creator, msg.IpfsId, msg.PublicIp, msg.Stake)

	found, err := ms.k.Workers.Has(ctx, msg.Creator)
	if err != nil {
		audioStemLogger.Logger.Error("Worker exists?: %s", err.Error())
		return &audioStem.MsgAddWorkerResponse{Ok: false, Message: err.Error()}, err
	}

	if found {
		audioStemLogger.Logger.Error("Worker %v already exists.", msg.Creator)
		error := sdkerrors.ErrAppConfig.Wrapf(audioStem.ErrWorkerAlreadyRegistered.Error(), "worker (%s) is already registered", msg.Creator)
		return &audioStem.MsgAddWorkerResponse{Ok: false, Message: error.Error()}, error
	}

	// we verify the staking amount if valid and at least equeal the min value
	params, _ := ms.k.Params.Get(ctx)
	if msg.Stake.Denom != params.MinWorkerStaking.Denom {
		error := sdkerrors.ErrAppConfig.Wrapf(audioStem.ErrWorkerIncorrectStake.Error(), "staked coin denom %s is not accepted", msg.Stake.Denom)
		audioStemLogger.Logger.Error(error.Error())
		return &audioStem.MsgAddWorkerResponse{Ok: false, Message: error.Error()}, error
	}

	if msg.Stake.Amount.LT(params.MinWorkerStaking.Amount) {
		error := sdkerrors.ErrAppConfig.Wrapf(audioStem.ErrWorkerIncorrectStake.Error(), "staked coin is not enought. Min value is %v", params.MinWorkerStaking.Amount)
		audioStemLogger.Logger.Error(error.Error())
		return &audioStem.MsgAddWorkerResponse{Ok: false, Message: error.Error()}, error
	}

	// we verify the account has enought balance to stack
	addr, _ := types.AccAddressFromBech32(msg.Creator)
	balance := ms.k.BankKeeper.GetBalance(ctx, addr, params.MinWorkerStaking.Denom)
	audioStemLogger.Logger.Debug("balance of %s [%s]: %s", msg.Creator, addr, balance)

	if balance.Amount.LT(params.MinWorkerStaking.Amount) {
		error := sdkerrors.ErrAppConfig.Wrapf(audioStem.ErrWorkerIncorrectStake.Error(), "not enought balance to stack. Min value is %v", params.MinWorkerStaking.Amount)
		audioStemLogger.Logger.Error(error.Error())
		return &audioStem.MsgAddWorkerResponse{Ok: false, Message: error.Error()}, error
	}

	// worker is not previously registered, so we move on
	reputation := audioStem.Worker_Reputation{Points: 0, Staked: &msg.Stake, Validations: 0, Solutions: 0, Winnings: types.NewCoin(params.MinWorkerStaking.Denom, math.NewInt(0))}
	worker := audioStem.Worker{Address: msg.Creator, Reputation: &reputation, Enabled: true, PublicIp: msg.PublicIp, IpfsId: msg.IpfsId}

	err = ms.k.Workers.Set(ctx, msg.Creator, worker)
	if err != nil {
		audioStemLogger.Logger.Error("Getting worker: %s", err.Error())
		return &audioStem.MsgAddWorkerResponse{Ok: false, Message: err.Error()}, err
	}

	// // we stack the coins in the module
	err = ms.k.BankKeeper.SendCoinsFromAccountToModule(ctx, addr, audioStem.ModuleName, types.NewCoins(msg.Stake))
	if err != nil {
		audioStemLogger.Logger.Error("Stacking worker's coins: %s", err.Error())
		return &audioStem.MsgAddWorkerResponse{Ok: false, Message: err.Error()}, err
	}

	return &audioStem.MsgAddWorkerResponse{Ok: true, Message: "Worker added correctly"}, nil
}

func (ms msgServer) SubscribeWorkerToTask(ctx context.Context, msg *audioStem.MsgSubscribeWorkerToTask) (*audioStem.MsgSubscribeWorkerToTaskResponse, error) {
	audioStemLogger.Logger.Info("SubscribeWorkerToTask - address: %s, taskId: %s, threadId: %s", msg.Address, msg.TaskId, msg.ThreadId)

	worker, err := ms.k.Workers.Get(ctx, msg.Address)
	if err != nil {
		audioStemLogger.Logger.Error("Getting Worker: %s", err.Error())
		return nil, err
	}

	if !worker.Enabled {
		audioStemLogger.Logger.Debug("Worker not enabled: %s", worker.String())
		return nil, sdkerrors.ErrAppConfig.Wrapf(audioStem.ErrWorkerNotAvailable.Error(), "worker (%s) it nos enabled or doesn't exists", msg.Address)
	}
	task, err := ms.k.AudioStemTasks.Get(ctx, msg.TaskId)
	if err != nil {
		audioStemLogger.Logger.Error("Getting task: %s", err.Error())
		return nil, err
	}
	if task.Completed {
		audioStemLogger.Logger.Debug("Task is completed: %s", task.String())
		return nil, sdkerrors.ErrAppConfig.Wrapf(audioStem.ErrWorkerTaskNotAvailable.Error(), "task (%s) is already completed. Can't subscribe worker", msg.TaskId)
	}

	// we get the params to get the MaxWorkersPerThread value
	params, _ := ms.k.Params.Get(ctx)
	for i, v := range task.Threads {
		if v.ThreadId == msg.ThreadId {
			if len(v.Workers) < int(params.MaxWorkersPerThread) && !v.Completed {

				if slices.Contains(v.Workers, worker.Address) {
					audioStemLogger.Logger.Info("worker %s is already working at thread %s, skipping...", worker.Address, v.ThreadId)
					return nil, nil
				}

				v.Workers = append(v.Workers, msg.Address)

				worker.CurrentTaskId = task.TaskId
				worker.CurrentThreadIndex = int32(i)
				ms.k.Workers.Set(ctx, msg.Address, worker)

				err := ms.k.AudioStemTasks.Set(ctx, task.TaskId, task)
				if err != nil {
					audioStemLogger.Logger.Error("error trying to update thread %s to in progress", v.ThreadId)
				}

				return &audioStem.MsgSubscribeWorkerToTaskResponse{ThreadId: v.ThreadId}, nil
			}
		}
	}
	return nil, nil
}

func (ms msgServer) ProposeSolution(ctx context.Context, msg *audioStem.MsgProposeSolution) (*audioStem.MsgProposeSolutionResponse, error) {
	audioStemLogger.Logger.Info("ProposeSolution - creator: %s, taskId: %s, threadId: %s, publicKey: %s, signatures: %s", msg.Creator, msg.TaskId, msg.ThreadId, msg.PublicKey, msg.Signatures)

	// creator of the solution must be a valid worker
	worker, err := ms.k.Workers.Get(ctx, msg.Creator)
	if err != nil {
		audioStemLogger.Logger.Error("Getting Worker: %s", err.Error())
		return nil, err
	}

	if !worker.Enabled {
		audioStemLogger.Logger.Error("workers %s is not enabled to propose a solution", msg.Creator)
		return nil, sdkerrors.ErrAppConfig.Wrapf(audioStem.ErrInvalidSolution.Error(), "workers %s is not enabled to propose a solution", msg.Creator)
	}

	task, err := ms.k.AudioStemTasks.Get(ctx, msg.TaskId)
	if err != nil {
		audioStemLogger.Logger.Error("Getting Task: %s", err.Error())
		return nil, err
	}

	// task must exists and be in progress
	if task.Completed {
		audioStemLogger.Logger.Error("Task %s is not valid to accept solutions", msg.TaskId)
		return nil, sdkerrors.ErrAppConfig.Wrapf(audioStem.ErrInvalidSolution.Error(), "Task %s is not valid to accept solutions", msg.TaskId)
	}

	for i, v := range task.Threads {
		// TODO threads might be better as map instead of slice
		if v.ThreadId == msg.ThreadId {
			if v.Solution != nil {
				audioStemLogger.Logger.Error("thread %s already has a solution", msg.ThreadId)
				return nil, sdkerrors.ErrAppConfig.Wrapf(audioStem.ErrInvalidSolution.Error(), "thread %s already has a solution", msg.ThreadId)
			}
			// worker must be a valid registered worker in the thread with a solution
			if !slices.Contains(v.Workers, msg.Creator) {
				audioStemLogger.Logger.Error("Worker %s is not valid at thread %s", msg.Creator, msg.ThreadId)
				return nil, sdkerrors.ErrAppConfig.Wrapf(audioStem.ErrInvalidSolution.Error(), "Worker %s is not valid at thread %s", msg.Creator, msg.ThreadId)
			}

			// solution len must be equal to the frames generated
			if len(msg.Signatures) != (int(v.EndFrame) - int(v.StartFrame) + 1) {
				audioStemLogger.Logger.Error("amount of files in solution is incorrect, %v ", len(msg.Signatures))
				return nil, sdkerrors.ErrAppConfig.Wrapf(audioStem.ErrInvalidSolution.Error(), "amount of files in solution is incorrect, %v ", len(msg.Signatures))
			}

			// we have passed all validations, lets add the solution to the thread
			var frames []*audioStem.AudioStemThread_Frame

			for _, val := range msg.Signatures {
				parts := strings.SplitN(val, "=", 2)

				if err != nil {
					audioStemLogger.Logger.Error("unable to decode signature from msg %s: %s", parts[1], err.Error())
					return nil, err
				}

				frame := audioStem.AudioStemThread_Frame{Filename: parts[0], Signature: parts[1]}
				frames = append(frames, &frame)
			}

			_, err := audioStemCrypto.DecodePublicKeyFromCLI(msg.PublicKey)

			if err != nil {
				audioStemLogger.Logger.Error("unable to decode publicKey from msg %s: %s", msg.PublicKey, err.Error())
				return nil, err
			}
			task.Threads[i].Solution = &audioStem.AudioStemThread_Solution{ProposedBy: msg.Creator, Frames: frames, PublicKey: msg.PublicKey}
			err = ms.k.AudioStemTasks.Set(ctx, msg.TaskId, task)

			if err != nil {
				audioStemLogger.Logger.Error("unable to propose solution %s", err.Error())
				return nil, err
			}
		}
	}

	return &audioStem.MsgProposeSolutionResponse{}, nil
}

// TODO Implement
func (ms msgServer) RevealSolution(ctx context.Context, msg *audioStem.MsgRevealSolution) (*audioStem.MsgRevealSolutionResponse, error) {
	audioStemLogger.Logger.Info("RevealSolution - creator: %s, taskId: %s, threadId: %s, frames: %s", msg.Creator, msg.TaskId, msg.ThreadId, msg.Frames)

	// Solution must be from a worker on the thread
	task, err := ms.k.AudioStemTasks.Get(ctx, msg.TaskId)

	if err != nil {
		audioStemLogger.Logger.Error("Getting task: %s", err.Error())
		return nil, err
	}

	worker, err := ms.k.Workers.Get(ctx, msg.Creator)

	if err != nil {
		audioStemLogger.Logger.Error("Getting Worker: %s", err.Error())
		return nil, err
	}

	if worker.CurrentTaskId != msg.TaskId {
		audioStemLogger.Logger.Error("worker is not working on task")
		return nil, sdkerrors.ErrAppConfig.Wrapf(audioStem.ErrInvalidVerification.Error(), "worker is not working on task")
	}

	if task.Completed {
		audioStemLogger.Logger.Error("task is already completed. No more validations accepted")
		return nil, sdkerrors.ErrAppConfig.Wrapf(audioStem.ErrInvalidVerification.Error(), "task is already completed. No more validations accepted")
	}

	thread := task.Threads[worker.CurrentThreadIndex]

	if thread.Solution.Accepted {
		audioStemLogger.Logger.Error("solution has already been accepted")
		return nil, sdkerrors.ErrAppConfig.Wrapf(audioStem.ErrInvalidVerification.Error(), "solution has already been accepted.")
	}

	if thread.Solution.ProposedBy != msg.Creator {
		audioStemLogger.Logger.Error("creator is not the winner.")
		return nil, sdkerrors.ErrAppConfig.Wrapf(audioStem.ErrInvalidVerification.Error(), "creator is not the winner.")
	}

	if thread.ThreadId != msg.ThreadId {
		audioStemLogger.Logger.Error("worker is not working on thread")
		return nil, sdkerrors.ErrAppConfig.Wrapf(audioStem.ErrInvalidVerification.Error(), "worker is not working on thread")
	}

	// this shouldn't happen.
	if !slices.Contains(thread.Workers, msg.Creator) {
		audioStemLogger.Logger.Error("worker is not working on thread")
		return nil, sdkerrors.ErrAppConfig.Wrapf(audioStem.ErrInvalidVerification.Error(), "worker is not working on thread")
	}

	// cids amount must be equal to the amount of frames
	if len(msg.Frames) != len(thread.Solution.Frames) {
		audioStemLogger.Logger.Error("invalid amount of frames for the solution")
		return nil, sdkerrors.ErrAppConfig.Wrapf(audioStem.ErrInvalidVerification.Error(), "invalid amount of cids for the solution")
	}

	solution := audioStem.FromCliToFrames(msg.Frames)
	for _, frame := range solution {
		idx := slices.IndexFunc(thread.Solution.Frames, func(f *audioStem.AudioStemThread_Frame) bool { return f.Filename == frame.Filename })
		if idx < 0 {
			audioStemLogger.Logger.Error("Unable to find frame with filename %s in solution", frame.String())
			return nil, sdkerrors.ErrAppConfig.Wrapf(audioStem.ErrInvalidVerification.Error(), "frame %s not found in solution", frame)
		}
		// we reveal the solution
		thread.Solution.Frames[idx].Cid = frame.Cid
		thread.Solution.Frames[idx].Hash = frame.Hash
	}

	// We verify all frames in the solution have a CID revealed
	for _, frame := range thread.Solution.Frames {
		if frame.Cid == "" || frame.Hash == "" {
			audioStemLogger.Logger.Error("Frame %s doesn't have a CID or Hash revelaed", frame.Filename)
			return nil, sdkerrors.ErrAppConfig.Wrapf(audioStem.ErrInvalidVerification.Error(), "Frame %s doesn't have a CID or Hash revelaed", frame.Filename)
		}
	}

	task.Threads[worker.CurrentThreadIndex] = thread
	ms.k.AudioStemTasks.Set(ctx, msg.TaskId, task)
	return nil, nil
}

func (ms msgServer) SubmitValidation(ctx context.Context, msg *audioStem.MsgSubmitValidation) (*audioStem.MsgSubmitValidationResponse, error) {
	audioStemLogger.Logger.Info("SubmitValidation - creator: %s, taskId: %s, threadId: %s, publicKey: %s, Signatures: %s", msg.Creator, msg.TaskId, msg.ThreadId, msg.PublicKey, msg.Signatures)

	// validation must be from a worker on the thread
	task, err := ms.k.AudioStemTasks.Get(ctx, msg.TaskId)

	if err != nil {
		audioStemLogger.Logger.Error("Getting Task: %s", err.Error())
		return nil, err
	}

	worker, err := ms.k.Workers.Get(ctx, msg.Creator)

	if err != nil {
		audioStemLogger.Logger.Error("Getting Worker: %s", err.Error())
		return nil, err
	}

	if !worker.Enabled {
		audioStemLogger.Logger.Error("worker is not allowed to validate solutions")
		return nil, sdkerrors.ErrAppConfig.Wrapf(audioStem.ErrInvalidVerification.Error(), "worker is not allowed to validate solutions")
	}

	if worker.CurrentTaskId != msg.TaskId {
		audioStemLogger.Logger.Error("worker is not working on task")
		return nil, sdkerrors.ErrAppConfig.Wrapf(audioStem.ErrInvalidVerification.Error(), "worker is not working on task")
	}

	if task.Completed {
		audioStemLogger.Logger.Error("task is already completed. No more validations accepted")
		return nil, sdkerrors.ErrAppConfig.Wrapf(audioStem.ErrInvalidVerification.Error(), "task is already completed. No more validations accepted")
	}

	thread := task.Threads[worker.CurrentThreadIndex]
	if thread.ThreadId != msg.ThreadId {
		audioStemLogger.Logger.Error("worker is not working on thread")
		return nil, sdkerrors.ErrAppConfig.Wrapf(audioStem.ErrInvalidVerification.Error(), "worker is not working on thread")
	}

	// this shouldn't happen.
	if !slices.Contains(thread.Workers, msg.Creator) {
		audioStemLogger.Logger.Error("worker is not working on thread")
		return nil, sdkerrors.ErrAppConfig.Wrapf(audioStem.ErrInvalidVerification.Error(), "worker is not working on thread")
	}

	var frames []*audioStem.AudioStemThread_Frame
	for _, signatures := range msg.Signatures {
		parts := strings.SplitN(signatures, "=", 2)

		frame := audioStem.AudioStemThread_Frame{Filename: parts[0], Signature: parts[1]}
		frames = append(frames, &frame)
	}

	validation := audioStem.AudioStemThread_Validation{Validator: msg.Creator, IsReverse: thread.IsReverse(worker.Address), Frames: frames, PublicKey: msg.PublicKey}
	task.Threads[worker.CurrentThreadIndex].Validations = append(thread.Validations, &validation)
	ms.k.AudioStemTasks.Set(ctx, msg.TaskId, task)

	// we release the worker since there is nothing else for him to do on this thread
	if worker.Address != thread.Solution.ProposedBy {
		worker.ReleaseValidator()
		ms.k.Workers.Set(ctx, msg.Creator, worker)
	}

	return &audioStem.MsgSubmitValidationResponse{}, nil
}

func (ms msgServer) SubmitSolution(ctx context.Context, msg *audioStem.MsgSubmitSolution) (*audioStem.MsgSubmitSolutionResponse, error) {
	audioStemLogger.Logger.Info("SubmitSolution - creator: %s, taskId: %s, threadId: %s, Dir: %s, AverageRenderSeconds: %v", msg.Creator, msg.TaskId, msg.ThreadId, msg.Dir, msg.AverageRenderSeconds)

	task, err := ms.k.AudioStemTasks.Get(ctx, msg.TaskId)
	if err != nil {
		audioStemLogger.Logger.Error("Getting Task: %s", err.Error())
		return nil, sdkerrors.ErrAppConfig.Wrapf(audioStem.ErrInvalidSolution.Error(), "provided task doesn't exists")
	}
	for i, thread := range task.Threads {
		if thread.ThreadId == msg.ThreadId {

			if thread.Solution.ProposedBy != msg.Creator {
				error := sdkerrors.ErrAppConfig.Wrapf(audioStem.ErrInvalidSolution.Error(), "only the provider of the solution can upload it")
				audioStemLogger.Logger.Error(error.Error())
				return nil, error
			}

			// we make sure ipfs is running
			ipfs.EnsureIPFSRunning()

			// we verify the solution
			// err := thread.VerifySubmittedSolution(msg.Dir)
			// if err != nil {
			// 	return nil, sdkerrors.ErrAppConfig.Wrapf(audioStem.ErrInvalidSolution.Error(), "submited solution is incorrect")
			// }

			// solution is verified so we pay the winner
			addr, _ := types.AccAddressFromBech32(msg.Creator)
			payment := task.GetWinnerReward()
			ms.k.BankKeeper.SendCoinsFromModuleToAccount(ctx, audioStem.ModuleName, addr, types.NewCoins(payment))
			task.Threads[i].Solution.Dir = msg.Dir
			task.Threads[i].AverageRenderSeconds = msg.AverageRenderSeconds
			task.Threads[i].Completed = true
			ms.k.AudioStemTasks.Set(ctx, msg.TaskId, task)

			// should we pay here the validators?
			// TODO Implement

			// a worker which didn't submit a validation might still be working on this task
			// we release them
			for _, val := range thread.Workers {
				worker, _ := ms.k.Workers.Get(ctx, val)
				if task.TaskId == worker.CurrentTaskId && thread.ThreadId == task.Threads[worker.CurrentThreadIndex].ThreadId {
					// this worker is still active but work is completed. we release him
					worker.CurrentTaskId = ""
					worker.CurrentThreadIndex = 0
					ms.k.Workers.Set(ctx, worker.Address, worker)
				}
			}

			// we increase the reputation of the winner
			worker, _ := ms.k.Workers.Get(ctx, msg.Creator)
			worker.DeclareWinner(payment)
			// we added this duration to the slice of average durations.
			worker.Reputation.RenderDurations = append(worker.Reputation.RenderDurations, msg.AverageRenderSeconds)
			ms.k.Workers.Set(ctx, msg.Creator, worker)
			break
		}
	}
	return &audioStem.MsgSubmitSolutionResponse{}, nil
}
