import { GeneratedType, OfflineSigner, Registry } from "@cosmjs/proto-signing"
import {
    defaultRegistryTypes,
    DeliverTxResponse,
    QueryClient,
    SigningStargateClient,
    SigningStargateClientOptions,
    StdFee,
} from "@cosmjs/stargate"
import { Tendermint34Client } from "@cosmjs/tendermint-rpc"
import Long from "long"
import { AudioStemExtension, setupAudioStemExtension } from "./modules/queries"
import {
    audioStemTypes,
    MsgCreateAudioStemTaskEncodeObject,
    typeUrlMsgCreateAudioStemTask,
} from "./modules/messages"

export const audioStemDefaultRegistryTypes: ReadonlyArray<[string, GeneratedType]> = [
    ...defaultRegistryTypes,
    ...audioStemTypes,
]

function createDefaultRegistry(): Registry {
    const registry = new Registry(defaultRegistryTypes)
    registry.register(audioStemTypes[0][0],audioStemTypes[0][1]  );
    registry.register(audioStemTypes[1][0],audioStemTypes[1][1]  );
    return registry
}

export class AudioStemSigningStargateClient extends SigningStargateClient {
    public readonly checkersQueryClient: AudioStemExtension | undefined

    public static async connectWithSigner(
        endpoint: string,
        signer: OfflineSigner,
        options: SigningStargateClientOptions = {},
    ): Promise<AudioStemSigningStargateClient> {
        const tmClient = await Tendermint34Client.connect(endpoint)
        return new AudioStemSigningStargateClient(tmClient, signer, {
            registry: createDefaultRegistry(),
            ...options,
        })
    }

    protected constructor(
        tmClient: Tendermint34Client | undefined,
        signer: OfflineSigner,
        options: SigningStargateClientOptions,
    ) {
        super(tmClient, signer, options)
        if (tmClient) {
            this.checkersQueryClient = QueryClient.withExtensions(tmClient, setupAudioStemExtension)
        }
    }

    public async createAudioStemTask(
        creator: string,
        cid: string,
        startFrame: number,
        endFrame: number,
        threads: number,
        reward: Long,
        fee: number | StdFee | "auto"
    ): Promise<DeliverTxResponse> {
        const createMsg: MsgCreateAudioStemTaskEncodeObject = {
            typeUrl: typeUrlMsgCreateAudioStemTask,
            value: {
                creator: creator,
                 cid: cid,
                startFrame: startFrame,
                endFrame: endFrame,
                threads: threads,
                reward: reward,
                
            },
        }
        return this.signAndBroadcast(creator, [createMsg],fee)
    }

    
}