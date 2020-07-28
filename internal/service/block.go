package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-kratos/kratos/pkg/log"
	"github.com/social-network/scale.go/types"
	"github.com/social-network/subscan/model"
	"github.com/social-network/subscan/util"
	"github.com/social-network/subscan/util/address"
	"github.com/social-network/substrate-api-rpc"
	"github.com/social-network/substrate-api-rpc/metadata"
	"github.com/social-network/substrate-api-rpc/rpc"
	"github.com/social-network/substrate-api-rpc/storage"
)

func (s *Service) CreateChainBlock(hash string, block *rpc.Block, event string, spec int, finalized bool) (err error) {
	var (
		decodeExtrinsics []map[string]interface{}
		decodeEvent      interface{}
		logs             []storage.DecoderLog
		validator        string
	)
	c := context.TODO()

	blockNum := util.StringToInt(util.HexToNumStr(block.Header.Number))

	metadataInstant := s.getMetadataInstant(spec)

	// Extrinsic
	decodeExtrinsics, err = substrate.DecodeExtrinsic(block.Extrinsics, metadataInstant, spec)
	if err != nil {
		log.Error("%v", err)
	}

	// event
	if err == nil {
		decodeEvent, err = substrate.DecodeEvent(event, metadataInstant, spec)
		if err != nil {
			log.Error("%v", err)
		}
	}

	// log
	if err == nil {
		logs, err = substrate.DecodeLogDigest(block.Header.Digest.Logs)
		if err != nil {
			log.Error("%v", err)
		}
	}

	txn := s.dao.DbBegin()
	defer s.dao.DbRollback(txn)

	var e []model.ChainEvent
	util.UnmarshalToAnything(&e, decodeEvent)

	eventMap := s.checkoutExtrinsicEvents(e, blockNum)

	cb := model.ChainBlock{
		Hash:           hash,
		BlockNum:       blockNum,
		ParentHash:     block.Header.ParentHash,
		StateRoot:      block.Header.StateRoot,
		ExtrinsicsRoot: block.Header.ExtrinsicsRoot,
		Logs:           util.InterfaceToString(block.Header.Digest.Logs),
		Extrinsics:     util.InterfaceToString(block.Extrinsics),
		Event:          event,
		SpecVersion:    spec,
		Finalized:      finalized,
	}

	extrinsicsCount, blockTimestamp, extrinsicHash, extrinsicFee, err := s.createExtrinsic(c, txn, &cb, block.Extrinsics, decodeExtrinsics, eventMap, finalized, spec)
	if err != nil {
		return err
	}
	cb.BlockTimestamp = blockTimestamp
	eventCount, err := s.AddEvent(c, txn, &cb, e, extrinsicHash, finalized, spec, extrinsicFee)
	if err != nil {
		return err
	}

	if validator, err = s.EmitLog(c, txn, hash, blockNum, logs, finalized); err != nil {
		return err
	}

	cb.Validator = validator
	cb.CodecError = validator == ""
	cb.ExtrinsicsCount = extrinsicsCount
	cb.EventCount = eventCount

	if err = s.dao.CreateBlock(txn, &cb); err == nil {
		s.dao.DbCommit(txn)
	}
	return err
}

func (s *Service) UpdateBlockData(block *model.ChainBlock, finalized bool) (err error) {
	c := context.TODO()

	var (
		decodeEvent      interface{}
		encodeExtrinsics []string
		decodeExtrinsics []map[string]interface{}
	)

	_ = json.Unmarshal([]byte(block.Extrinsics), &encodeExtrinsics)

	spec := block.SpecVersion

	metadataInstant := s.getMetadataInstant(spec)

	// Event
	decodeEvent, err = substrate.DecodeEvent(block.Event, metadataInstant, spec)
	if err != nil {
		fmt.Println("ERR: Decode Event get error ", err)
		return
	}

	// Extrinsic
	decodeExtrinsics, err = substrate.DecodeExtrinsic(encodeExtrinsics, metadataInstant, spec)
	if err != nil {
		fmt.Println("ERR: Decode Extrinsic get error ", err)
		return
	}

	// Log
	var rawList []string
	_ = json.Unmarshal([]byte(block.Logs), &rawList)
	logs, err := substrate.DecodeLogDigest(rawList)
	if err != nil {
		fmt.Println("ERR: Decode Logs get error ", err)
		return
	}

	var e []model.ChainEvent
	util.UnmarshalToAnything(&e, decodeEvent)
	eventMap := s.checkoutExtrinsicEvents(e, block.BlockNum)

	txn := s.dao.DbBegin()
	defer s.dao.DbRollback(txn)

	extrinsicsCount, blockTimestamp, extrinsicHash, extrinsicFee, err := s.createExtrinsic(c, txn, block, encodeExtrinsics, decodeExtrinsics, eventMap, finalized, spec)
	if err != nil {
		return err
	}
	block.BlockTimestamp = blockTimestamp

	eventCount, err := s.AddEvent(c, txn, block, e, extrinsicHash, finalized, spec, extrinsicFee)
	if err != nil {
		return err
	}

	validator, err := s.EmitLog(c, txn, block.Hash, block.BlockNum, logs, finalized)
	if err != nil {
		return err
	}

	if err = s.dao.UpdateEventAndExtrinsic(txn, block, eventCount, extrinsicsCount, blockTimestamp, validator, validator == "", finalized); err != nil {
		return
	}

	s.dao.DbCommit(txn)
	return
}

func (s *Service) checkoutExtrinsicEvents(e []model.ChainEvent, blockNumInt int) map[string][]model.ChainEvent {
	eventMap := make(map[string][]model.ChainEvent)
	for _, event := range e {
		extrinsicIndex := fmt.Sprintf("%d-%d", blockNumInt, event.ExtrinsicIdx)
		eventMap[extrinsicIndex] = append(eventMap[extrinsicIndex], event)
	}
	return eventMap
}

func (s *Service) GetCurrentRuntimeSpecVersion(blockNum int) int {
	if util.CurrentRuntimeSpecVersion != 0 {
		return util.CurrentRuntimeSpecVersion
	}
	if block := s.dao.GetNearBlock(blockNum); block != nil {
		return block.SpecVersion
	}
	return -1
}

func (s *Service) getMetadataInstant(spec int) *types.MetadataStruct {
	metadataInstant, ok := metadata.RuntimeMetadata[spec]
	if !ok {
		metadataInstant = metadata.Process(s.dao.RuntimeVersionRaw(spec))
	}
	return metadataInstant
}

func (s *Service) GetExtrinsicList(page, row int, order string, query ...string) ([]*model.ChainExtrinsicJson, int) {
	c := context.TODO()
	list, count := s.dao.GetExtrinsicList(c, page, row, order, query...)
	var ejs []*model.ChainExtrinsicJson
	for _, extrinsic := range list {
		ejs = append(ejs, s.dao.ExtrinsicsAsJson(&extrinsic))
	}
	return ejs, count
}

func (s *Service) GetBlocksSampleByNums(page, row int) []model.SampleBlockJson {
	c := context.TODO()
	var blockJson []model.SampleBlockJson
	blocks := s.dao.GetBlockList(page, row)
	for _, block := range blocks {
		bj := s.BlockAsSampleJson(c, &block)
		blockJson = append(blockJson, *bj)
	}
	return blockJson
}

func (s *Service) GetExtrinsicByIndex(index string) *model.ExtrinsicDetail {
	c := context.TODO()
	return s.dao.GetExtrinsicsDetailByIndex(c, index)
}

func (s *Service) GetExtrinsicDetailByHash(hash string) *model.ExtrinsicDetail {
	c := context.TODO()
	return s.dao.GetExtrinsicsDetailByHash(c, hash)
}

func (s *Service) GetExtrinsicByHash(hash string) *model.ChainExtrinsic {
	c := context.TODO()
	return s.dao.GetExtrinsicsByHash(c, hash)
}

func (s *Service) GetBlockByHashJson(hash string) *model.ChainBlockJson {
	c := context.TODO()
	block := s.dao.GetBlockByHash(c, hash)
	if block == nil {
		return nil
	}
	return s.dao.BlockAsJson(c, block)
}

func (s *Service) EventByIndex(index string) *model.ChainEvent {
	return s.dao.GetEventByIdx(index)
}

func (s *Service) GetBlockByNum(num int) *model.ChainBlockJson {
	c := context.TODO()
	block := s.dao.GetBlockByNum(num)
	if block == nil {
		return nil
	}
	return s.dao.BlockAsJson(c, block)
}

func (s *Service) GetBlockByHash(hash string) *model.ChainBlock {
	c := context.TODO()
	block := s.dao.GetBlockByHash(c, hash)
	if block == nil {
		return nil
	}
	return block
}

func (s *Service) BlockAsSampleJson(c context.Context, block *model.ChainBlock) *model.SampleBlockJson {
	b := model.SampleBlockJson{
		BlockNum:        block.BlockNum,
		BlockTimestamp:  block.BlockTimestamp,
		Hash:            block.Hash,
		EventCount:      block.EventCount,
		ExtrinsicsCount: block.ExtrinsicsCount,
		Validator:       address.SS58Address(block.Validator),
		Finalized:       block.Finalized,
	}
	return &b
}
