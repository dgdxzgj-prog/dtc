import { GeneratedType } from "@cosmjs/proto-signing";
import { MsgUpdateParams } from "./types/dtc/task/v1/tx";
import { MsgCreateClaimRecord } from "./types/dtc/task/v1/tx";
import { MsgUpdateClaimRecord } from "./types/dtc/task/v1/tx";
import { MsgDeleteClaimRecord } from "./types/dtc/task/v1/tx";
import { MsgClaimReward } from "./types/dtc/task/v1/tx";

const msgTypes: Array<[string, GeneratedType]>  = [
    ["/dtc.task.v1.MsgUpdateParams", MsgUpdateParams],
    ["/dtc.task.v1.MsgCreateClaimRecord", MsgCreateClaimRecord],
    ["/dtc.task.v1.MsgUpdateClaimRecord", MsgUpdateClaimRecord],
    ["/dtc.task.v1.MsgDeleteClaimRecord", MsgDeleteClaimRecord],
    ["/dtc.task.v1.MsgClaimReward", MsgClaimReward],
    
];

export { msgTypes }