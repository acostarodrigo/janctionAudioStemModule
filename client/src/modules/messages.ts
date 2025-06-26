import { EncodeObject, GeneratedType } from "@cosmjs/proto-signing"
import {
    MsgCreateAudioStemTask,
    MsgCreateAudioStemTaskResponse
} from "../types/generated/janction/audioStem/v1/tx"

export const typeUrlMsgCreateAudioStemTask = "/janction.audioStem.v1.MsgCreateAudioStemTask"
export const typeUrlMsgCreateAudioStemTaskResponse = "/janction.audioStem.v1.MsgCreateAudioStemTaskResponse"

export const audioStemTypes: ReadonlyArray<[string, GeneratedType]> = [
    [typeUrlMsgCreateAudioStemTask, MsgCreateAudioStemTask],
    [typeUrlMsgCreateAudioStemTaskResponse, MsgCreateAudioStemTaskResponse],
    
]

export interface MsgCreateAudioStemTaskEncodeObject extends EncodeObject {
    readonly typeUrl: "/janction.audioStem.v1.MsgCreateAudioStemTask"
    readonly value: Partial<MsgCreateAudioStemTask>
}

export function isMsgCreateAudioStemTaskEncodeObject(
    encodeObject: EncodeObject,
): encodeObject is MsgCreateAudioStemTaskEncodeObject {
    return (encodeObject as MsgCreateAudioStemTaskEncodeObject).typeUrl === typeUrlMsgCreateAudioStemTask
}

export interface MsgCreateAudioStemTaskResponseEncodeObject extends EncodeObject {
    readonly typeUrl: "/janction.audioStem.v1.MsgCreateAudioStemTaskResponse"
    readonly value: Partial<MsgCreateAudioStemTaskResponse>
}

export function isMsgCreateAudioStemTaskResponseEncodeObject(
    encodeObject: EncodeObject,
): encodeObject is MsgCreateAudioStemTaskResponseEncodeObject {
    return (encodeObject as MsgCreateAudioStemTaskResponseEncodeObject).typeUrl === typeUrlMsgCreateAudioStemTaskResponse
}

