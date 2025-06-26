import { QueryClient, StargateClient, StargateClientOptions } from "@cosmjs/stargate"
import { Tendermint34Client } from "@cosmjs/tendermint-rpc"
import { AudioStemExtension, setupAudioStemExtension } from "./modules/queries"

export class AudioStemStargateClient extends StargateClient {
    public readonly audioStemQueryClient: AudioStemExtension | undefined

    public static async connect(
        endpoint: string,
        options?: StargateClientOptions,
    ): Promise<AudioStemStargateClient> {
        const tmClient = await Tendermint34Client.connect(endpoint)
        return new AudioStemStargateClient(tmClient, options)
    }

    protected constructor(tmClient: Tendermint34Client | undefined, options: StargateClientOptions = {}) {
        super(tmClient, options)
        if (tmClient) {
            this.audioStemQueryClient = QueryClient.withExtensions(tmClient, setupAudioStemExtension)
        }
    }
}