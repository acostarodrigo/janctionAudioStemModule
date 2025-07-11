package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	storetypes "cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/janction/audioStem"
	"github.com/janction/audioStem/db"
)

type Keeper struct {
	cdc          codec.BinaryCodec
	addressCodec address.Codec
	BankKeeper   bankkeeper.BaseKeeper

	// authority is the address capable of executing a MsgUpdateParams and other authority-gated message.
	// typically, this should be the x/gov module account.
	authority string

	// state management
	Schema            collections.Schema
	Params            collections.Item[audioStem.Params]
	AudioStemTaskInfo collections.Item[audioStem.AudioStemTaskInfo]
	AudioStemTasks    collections.Map[string, audioStem.AudioStemTask]
	Workers           collections.Map[string, audioStem.Worker]
	Configuration     VideoConfiguration
	DB                db.DB
}

// NewKeeper creates a new Keeper instance
func NewKeeper(cdc codec.BinaryCodec, addressCodec address.Codec, storeService storetypes.KVStoreService, authority string, path string, bankKeeper bankkeeper.BaseKeeper) Keeper {
	if _, err := addressCodec.StringToBytes(authority); err != nil {
		panic(fmt.Errorf("invalid authority address: %w", err))
	}

	// we initialize the database
	db, err := db.Init(path)
	if err != nil {
		panic(err)
	}

	config, _ := GetAudioStemConfiguration(path)

	sb := collections.NewSchemaBuilder(storeService)
	k := Keeper{
		cdc:               cdc,
		addressCodec:      addressCodec,
		authority:         authority,
		Params:            collections.NewItem(sb, audioStem.ParamsKey, "params", codec.CollValue[audioStem.Params](cdc)),
		AudioStemTaskInfo: collections.NewItem(sb, audioStem.TaskInfoKey, "audioStemtaskInfo", codec.CollValue[audioStem.AudioStemTaskInfo](cdc)),
		AudioStemTasks:    collections.NewMap(sb, audioStem.AudioStemTaskKey, "audioStemTasks", collections.StringKey, codec.CollValue[audioStem.AudioStemTask](cdc)),
		Workers:           collections.NewMap(sb, audioStem.WorkerKey, "audioStemWorkers", collections.StringKey, codec.CollValue[audioStem.Worker](cdc)),
		Configuration:     *config,
		DB:                *db,
		BankKeeper:        bankKeeper,
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}

	k.Schema = schema

	return k
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}
