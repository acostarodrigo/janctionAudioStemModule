import { expect } from "chai"
import { config } from "dotenv"
import _ from "../../environment"
import { AudioStemStargateClient } from "../../src/stargateClient"
import { AudioStemExtension } from "../../src/modules/queries"

config()

describe("AudioStem", function () {
    let client: AudioStemStargateClient, audioStem: AudioStemExtension["audioStem"]

    before("create client", async function () {
        client = await AudioStemStargateClient.connect(process.env.RPC_URL)
        audioStem = client.audioStemQueryClient!.audioStem
    })

    it("can get audioStemtask", async function () {
        const task = await audioStem.GetAudioStemTask('1')
        expect(task.cid).to.be.equal("QmYC32RNLAMPRa8RGWEEHJWMcrnMzJ2Hq8xByupeFPUNtn")
    })


})