package audioStem

import (
	types "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterInterfaces registers the interfaces types with the interface registry.
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreateAudioStemTask{},
		&MsgAddWorker{},
		&MsgSubscribeWorkerToTask{},
		&MsgProposeSolution{},
		&MsgSubmitValidation{},
		&MsgRevealSolution{},
		&MsgSubmitSolution{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
