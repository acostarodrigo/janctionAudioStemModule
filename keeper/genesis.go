package keeper

import (
	"context"

	"github.com/janction/audioStem"
)

// InitGenesis initializes the module state from a genesis state.
func (k *Keeper) InitGenesis(ctx context.Context, data *audioStem.GenesisState) error {
	if err := k.Params.Set(ctx, data.Params); err != nil {
		return err
	}

	if err := k.AudioStemTaskInfo.Set(ctx, data.AudioStemTaskInfo); err != nil {
		return err
	}

	return nil
}

// ExportGenesis exports the module state to a genesis state.
func (k *Keeper) ExportGenesis(ctx context.Context) (*audioStem.GenesisState, error) {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	return &audioStem.GenesisState{
		Params:            params,
		AudioStemTaskList: []audioStem.IndexedAudioStemTask{},
		AudioStemTaskInfo: audioStem.AudioStemTaskInfo{NextId: 1},
	}, nil
}
