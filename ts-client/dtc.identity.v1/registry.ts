import { GeneratedType } from "@cosmjs/proto-signing";
import { MsgUpdateParams } from "./types/dtc/identity/v1/tx";
import { MsgCreateDidDocument } from "./types/dtc/identity/v1/tx";
import { MsgUpdateDidDocument } from "./types/dtc/identity/v1/tx";
import { MsgDeleteDidDocument } from "./types/dtc/identity/v1/tx";

const msgTypes: Array<[string, GeneratedType]>  = [
    ["/dtc.identity.v1.MsgUpdateParams", MsgUpdateParams],
    ["/dtc.identity.v1.MsgCreateDidDocument", MsgCreateDidDocument],
    ["/dtc.identity.v1.MsgUpdateDidDocument", MsgUpdateDidDocument],
    ["/dtc.identity.v1.MsgDeleteDidDocument", MsgDeleteDidDocument],
    
];

export { msgTypes }