package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/social-network/subscan-plugin/storage"
	"github.com/social-network/netscan/internal/dao"
	"github.com/social-network/netscan/internal/service/transaction"
	"github.com/social-network/netscan/model"
	"github.com/social-network/netscan/plugins"
	"github.com/social-network/netscan/util"
	"github.com/social-network/netscan/util/address"
	"github.com/shopspring/decimal"
	"strings"
)

func (s *Service) createExtrinsic(c context.Context,
	txn *dao.GormDB,
	block *model.ChainBlock,
	encodeExtrinsics []string,
	decodeExtrinsics []map[string]interface{},
	eventMap map[string][]model.ChainEvent,
	finalized bool,
	spec int,
) (int, int, map[string]string, map[string]decimal.Decimal, error) {

	var (
		blockTimestamp int
		e              []model.ChainExtrinsic
		err            error
	)
	extrinsicFee := make(map[string]decimal.Decimal)

	eb, _ := json.Marshal(decodeExtrinsics)
	_ = json.Unmarshal(eb, &e)

	hash := make(map[string]string)

	s.dao.DropExtrinsicNotFinalizedData(c, block.BlockNum, finalized)

	for index, extrinsic := range e {
		extrinsic.CallModule = strings.ToLower(extrinsic.CallModule)
		extrinsic.BlockNum = block.BlockNum
		extrinsic.ExtrinsicIndex = fmt.Sprintf("%d-%d", extrinsic.BlockNum, index)
		extrinsic.Success = s.getExtrinsicSuccess(eventMap[extrinsic.ExtrinsicIndex])
		extrinsic.Finalized = finalized

		s.getTimestamp(&extrinsic)

		if extrinsic.ExtrinsicHash != "" {
			fee, _ := transaction.GetExtrinsicFee(encodeExtrinsics[index])
			extrinsic.Fee = fee
			extrinsicFee[extrinsic.ExtrinsicIndex] = fee
			hash[extrinsic.ExtrinsicIndex] = extrinsic.ExtrinsicHash
		}

		if err = s.dao.CreateExtrinsic(c, txn, &extrinsic); err == nil {
			go s.afterExtrinsic(block, &extrinsic, eventMap[extrinsic.ExtrinsicIndex])
		} else {
			return 0, 0, nil, nil, err
		}
	}
	return len(e), blockTimestamp, hash, extrinsicFee, err
}

func (s *Service) ExtrinsicsAsJson(e *model.ChainExtrinsic) *model.ChainExtrinsicJson {
	ej := &model.ChainExtrinsicJson{
		BlockNum:           e.BlockNum,
		BlockTimestamp:     e.BlockTimestamp,
		ExtrinsicIndex:     e.ExtrinsicIndex,
		ExtrinsicHash:      e.ExtrinsicHash,
		Success:            e.Success,
		CallModule:         e.CallModule,
		CallModuleFunction: e.CallModuleFunction,
		Params:             util.InterfaceToString(e.Params),
		AccountId:          address.SS58Address(e.AccountId),
		Signature:          e.Signature,
		Nonce:              e.Nonce,
		Fee:                e.Fee,
	}
	var paramsInstant []model.ExtrinsicParam
	if err := json.Unmarshal([]byte(ej.Params), &paramsInstant); err != nil {
		for pi, param := range paramsInstant {
			if paramsInstant[pi].Type == "Address" {
				paramsInstant[pi].Value = address.SS58Address(param.Value.(string))
			}
		}
		bp, _ := json.Marshal(paramsInstant)
		ej.Params = string(bp)
	}
	return ej
}

func (s *Service) getTimestamp(extrinsic *model.ChainExtrinsic) {
	if extrinsic.CallModule != "timestamp" {
		return
	}

	var paramsInstant []model.ExtrinsicParam
	util.UnmarshalToAnything(&paramsInstant, extrinsic.Params)

	for _, p := range paramsInstant {
		if p.Name == "now" {
			extrinsic.BlockTimestamp = util.IntFromInterface(p.Value)
		}
	}
}

func (s *Service) getExtrinsicSuccess(e []model.ChainEvent) bool {
	for _, event := range e {
		if strings.EqualFold(event.ModuleId, "system") {
			return strings.EqualFold(event.EventId, "ExtrinsicFailed")
		}
	}
	return true
}

func (s *Service) afterExtrinsic(block *model.ChainBlock, extrinsic *model.ChainExtrinsic, events []model.ChainEvent) {
	block.BlockTimestamp = extrinsic.BlockTimestamp
	pBlock := block.AsPluginBlock()
	pExtrinsic := extrinsic.AsPluginExtrinsic()

	var pEvents []storage.Event
	for _, event := range events {
		pEvents = append(pEvents, *event.AsPluginEvent())
	}

	for _, plugin := range plugins.RegisteredPlugins {
		_ = plugin.ProcessExtrinsic(pBlock, pExtrinsic, pEvents)
	}
}
