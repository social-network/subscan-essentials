package service

import (
	"context"
	"fmt"
	"github.com/social-network/subscan/internal/dao"
	"github.com/social-network/subscan/model"
	"github.com/social-network/subscan/plugins"
	"github.com/social-network/subscan/util"
	"github.com/shopspring/decimal"
	"strings"
)

func (s *Service) AddEvent(
	c context.Context,
	txn *dao.GormDB,
	block *model.ChainBlock,
	e []model.ChainEvent,
	hashMap map[string]string,
	finalized bool,
	spec int,
	feeMap map[string]decimal.Decimal,
) (eventCount int, err error) {

	s.dao.DropEventNotFinalizedData(block.BlockNum, finalized)
	for _, event := range e {
		event.ModuleId = strings.ToLower(event.ModuleId)
		event.ExtrinsicHash = hashMap[fmt.Sprintf("%d-%d", block.BlockNum, event.ExtrinsicIdx)]
		event.EventIndex = fmt.Sprintf("%d-%d", block.BlockNum, event.ExtrinsicIdx)
		event.Finalized = finalized
		event.BlockNum = block.BlockNum

		if err = s.dao.CreateEvent(c, txn, &event); err == nil {
			go s.afterEvent(block, &event, feeMap[event.EventIndex])
		} else {
			return 0, err
		}
		eventCount++
	}
	return eventCount, err
}

func (s *Service) GetEventList(page, row int, order string, where ...string) ([]model.ChainEventJson, int) {
	c := context.TODO()
	var result []model.ChainEventJson
	var blockNums []int

	list, count := s.dao.GetEventList(c, page, row, order, where...)
	for _, event := range list {
		blockNums = append(blockNums, event.BlockNum)
	}
	blockMap := s.dao.BlocksReverseByNum(blockNums)

	for _, event := range list {
		ej := model.ChainEventJson{
			ExtrinsicIdx:  event.ExtrinsicIdx,
			EventIndex:    event.EventIndex,
			BlockNum:      event.BlockNum,
			ModuleId:      event.ModuleId,
			EventId:       event.EventId,
			Params:        util.InterfaceToString(event.Params),
			EventIdx:      event.EventIdx,
			ExtrinsicHash: event.ExtrinsicHash,
		}
		if block, ok := blockMap[event.BlockNum]; ok {
			ej.BlockTimestamp = block.BlockTimestamp
		}
		result = append(result, ej)
	}
	return result, count
}

func (s *Service) afterEvent(block *model.ChainBlock, event *model.ChainEvent, fee decimal.Decimal) {
	pBlock := block.AsPluginBlock()
	pEvent := event.AsPluginEvent()
	for _, plugin := range plugins.RegisteredPlugins {
		_ = plugin.ProcessEvent(pBlock, pEvent, fee)
	}

}
