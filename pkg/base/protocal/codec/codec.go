/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package codec

import (
	"bytes"
)

import (
	"vimagination.zapto.org/byteio"
)

import (
	"github.com/transaction-mesh/starfish/pkg/base/protocal"
	"github.com/transaction-mesh/starfish/pkg/util/log"
)

type SerializerType byte

const (
	SEATA    = byte(0x1)
	PROTOBUF = byte(0x2)
	KRYO     = byte(0x4)
	FST      = byte(0x8)
)

type Encoder func(in interface{}) []byte

type Decoder func(in []byte) (interface{}, int)

func MessageEncoder(codecType byte, in interface{}) []byte {
	switch codecType {
	case SEATA:
		return StarfishEncoder(in)
	default:
		log.Errorf("not support codecType, %s", codecType)
		return nil
	}
}

func MessageDecoder(codecType byte, in []byte) (interface{}, int) {
	switch codecType {
	case SEATA:
		return StarfishDecoder(in)
	default:
		log.Errorf("not support codecType, %s", codecType)
		return nil, 0
	}
}

func StarfishEncoder(in interface{}) []byte {
	var result = make([]byte, 0)
	msg := in.(protocal.MessageTypeAware)
	typeCode := msg.GetTypeCode()
	encoder := getMessageEncoder(typeCode)

	typeC := uint16(typeCode)
	if encoder != nil {
		body := encoder(in)
		result = append(result, []byte{byte(typeC >> 8), byte(typeC)}...)
		result = append(result, body...)
	}
	return result
}

func StarfishDecoder(in []byte) (interface{}, int) {
	r := byteio.BigEndianReader{Reader: bytes.NewReader(in)}
	typeCode, _, _ := r.ReadInt16()

	decoder := getMessageDecoder(typeCode)
	if decoder != nil {
		return decoder(in[2:])
	}
	return nil, 0
}

func getMessageEncoder(typeCode int16) Encoder {
	switch typeCode {
	case protocal.TypeStarfishMerge:
		return MergedWarpMessageEncoder
	case protocal.TypeStarfishMergeResult:
		return MergeResultMessageEncoder
	case protocal.TypeRegClt:
		return RegisterTMRequestEncoder
	case protocal.TypeRegCltResult:
		return RegisterTMResponseEncoder
	case protocal.TypeRegRm:
		return RegisterRMRequestEncoder
	case protocal.TypeRegRmResult:
		return RegisterRMResponseEncoder
	case protocal.TypeBranchCommit:
		return BranchCommitRequestEncoder
	case protocal.TypeBranchRollback:
		return BranchRollbackRequestEncoder
	case protocal.TypeGlobalReport:
		return GlobalReportRequestEncoder
	default:
		var encoder Encoder
		encoder = getMergeRequestMessageEncoder(typeCode)
		if encoder != nil {
			return encoder
		}
		encoder = getMergeResponseMessageEncoder(typeCode)
		if encoder != nil {
			return encoder
		}
		log.Errorf("not support typeCode, %d", typeCode)
		return nil
	}
}

func getMergeRequestMessageEncoder(typeCode int16) Encoder {
	switch typeCode {
	case protocal.TypeGlobalBegin:
		return GlobalBeginRequestEncoder
	case protocal.TypeGlobalCommit:
		return GlobalCommitRequestEncoder
	case protocal.TypeGlobalRollback:
		return GlobalRollbackRequestEncoder
	case protocal.TypeGlobalStatus:
		return GlobalStatusRequestEncoder
	case protocal.TypeGlobalLockQuery:
		return GlobalLockQueryRequestEncoder
	case protocal.TypeBranchRegister:
		return BranchRegisterRequestEncoder
	case protocal.TypeBranchStatusReport:
		return BranchReportRequestEncoder
	case protocal.TypeRmDeleteUndolog:
		return UndoLogDeleteRequestEncoder
	case protocal.TypeGlobalReport:
		return GlobalReportRequestEncoder
	default:
		break
	}
	return nil
}

func getMergeResponseMessageEncoder(typeCode int16) Encoder {
	switch typeCode {
	case protocal.TypeGlobalBeginResult:
		return GlobalBeginResponseEncoder
	case protocal.TypeGlobalCommitResult:
		return GlobalCommitResponseEncoder
	case protocal.TypeGlobalRollbackResult:
		return GlobalRollbackResponseEncoder
	case protocal.TypeGlobalStatusResult:
		return GlobalStatusResponseEncoder
	case protocal.TypeGlobalLockQueryResult:
		return GlobalLockQueryResponseEncoder
	case protocal.TypeBranchRegisterResult:
		return BranchRegisterResponseEncoder
	case protocal.TypeBranchStatusReportResult:
		return BranchReportResponseEncoder
	case protocal.TypeBranchCommitResult:
		return BranchCommitResponseEncoder
	case protocal.TypeBranchRollbackResult:
		return BranchRollbackResponseEncoder
	case protocal.TypeGlobalReportResult:
		return GlobalReportResponseEncoder
	default:
		break
	}
	return nil
}

func getMessageDecoder(typeCode int16) Decoder {
	switch typeCode {
	case protocal.TypeStarfishMerge:
		return MergedWarpMessageDecoder
	case protocal.TypeStarfishMergeResult:
		return MergeResultMessageDecoder
	case protocal.TypeRegClt:
		return RegisterTMRequestDecoder
	case protocal.TypeRegCltResult:
		return RegisterTMResponseDecoder
	case protocal.TypeRegRm:
		return RegisterRMRequestDecoder
	case protocal.TypeRegRmResult:
		return RegisterRMResponseDecoder
	case protocal.TypeBranchCommit:
		return BranchCommitRequestDecoder
	case protocal.TypeBranchRollback:
		return BranchRollbackRequestDecoder
	case protocal.TypeGlobalReport:
		return GlobalReportRequestDecoder
	default:
		var Decoder Decoder
		Decoder = getMergeRequestMessageDecoder(typeCode)
		if Decoder != nil {
			return Decoder
		}
		Decoder = getMergeResponseMessageDecoder(typeCode)
		if Decoder != nil {
			return Decoder
		}
		log.Errorf("not support typeCode, %d", typeCode)
		return nil
	}
}

func getMergeRequestMessageDecoder(typeCode int16) Decoder {
	switch typeCode {
	case protocal.TypeGlobalBegin:
		return GlobalBeginRequestDecoder
	case protocal.TypeGlobalCommit:
		return GlobalCommitRequestDecoder
	case protocal.TypeGlobalRollback:
		return GlobalRollbackRequestDecoder
	case protocal.TypeGlobalStatus:
		return GlobalStatusRequestDecoder
	case protocal.TypeGlobalLockQuery:
		return GlobalLockQueryRequestDecoder
	case protocal.TypeBranchRegister:
		return BranchRegisterRequestDecoder
	case protocal.TypeBranchStatusReport:
		return BranchReportRequestDecoder
	case protocal.TypeRmDeleteUndolog:
		return UndoLogDeleteRequestDecoder
	case protocal.TypeGlobalReport:
		return GlobalReportRequestDecoder
	default:
		break
	}
	return nil
}

func getMergeResponseMessageDecoder(typeCode int16) Decoder {
	switch typeCode {
	case protocal.TypeGlobalBeginResult:
		return GlobalBeginResponseDecoder
	case protocal.TypeGlobalCommitResult:
		return GlobalCommitResponseDecoder
	case protocal.TypeGlobalRollbackResult:
		return GlobalRollbackResponseDecoder
	case protocal.TypeGlobalStatusResult:
		return GlobalStatusResponseDecoder
	case protocal.TypeGlobalLockQueryResult:
		return GlobalLockQueryResponseDecoder
	case protocal.TypeBranchRegisterResult:
		return BranchRegisterResponseDecoder
	case protocal.TypeBranchStatusReportResult:
		return BranchReportResponseDecoder
	case protocal.TypeBranchCommitResult:
		return BranchCommitResponseDecoder
	case protocal.TypeBranchRollbackResult:
		return BranchRollbackResponseDecoder
	case protocal.TypeGlobalReportResult:
		return GlobalReportResponseDecoder
	default:
		break
	}
	return nil
}
