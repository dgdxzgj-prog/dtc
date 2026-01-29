import { GeneratedType } from "@cosmjs/proto-signing";
import { MsgUpdateParams } from "./types/dtc/credit/v1/tx";
import { MsgMintCredit } from "./types/dtc/credit/v1/tx";
import { MsgSubmitDeathCertificate } from "./types/dtc/credit/v1/tx";

const msgTypes: Array<[string, GeneratedType]>  = [
    ["/dtc.credit.v1.MsgUpdateParams", MsgUpdateParams],
    ["/dtc.credit.v1.MsgMintCredit", MsgMintCredit],
    ["/dtc.credit.v1.MsgSubmitDeathCertificate", MsgSubmitDeathCertificate],
    
];

export { msgTypes }