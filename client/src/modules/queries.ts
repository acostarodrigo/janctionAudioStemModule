import { createProtobufRpcClient, QueryClient } from "@cosmjs/stargate"
import {AudioStemLogs, AudioStemTask, Worker} from '../types/generated/janction/audioStem/v1/types'
import {QueryClientImpl, QueryGetAudioStemLogsResponse, QueryGetWorkerResponse} from '../types/generated/janction/audioStem/v1/query'
import {QueryGetAudioStemTaskResponse} from '../../src/types/generated/janction/audioStem/v1/query'

export interface AudioStemExtension {
    readonly audioStem: {
        readonly GetAudioStemTask: (
            index: string
        ) => Promise<AudioStemTask | undefined>;

        readonly GetAudioStemLog: (
            threadId: string
        ) => Promise<AudioStemLogs | undefined>;
        
        readonly GetWorker: (
            worker: string
        ) => Promise<Worker | undefined>;
    };
}

export function setupAudioStemExtension(base: QueryClient): AudioStemExtension {
    const rpc = createProtobufRpcClient(base);
    const queryService = new QueryClientImpl(rpc);

    return {
        audioStem: {
            GetAudioStemTask: async (index: string): Promise<AudioStemTask | undefined> => {
                const response: QueryGetAudioStemTaskResponse = await queryService.GetAudioStemTask({
                    index: index,
                });
                return response.audioStemTask;
            },

            GetAudioStemLog: async (threadId: string): Promise<AudioStemLogs | undefined> => {
                const response: QueryGetAudioStemLogsResponse = await queryService.GetAudioStemLogs({
                    threadId: threadId,
                });
                return response.audioStemLogs;
            },

            GetWorker: async (worker: string): Promise<Worker | undefined> => {
                const response: QueryGetWorkerResponse = await queryService.GetWorker({
                    worker: worker,
                });
                return response.worker;
            },
        },
    };
}
