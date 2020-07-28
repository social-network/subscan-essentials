package service

import (
	"github.com/social-network/netscan/util"
	"github.com/social-network/substrate-api-rpc/rpc"
	"github.com/social-network/substrate-api-rpc/websocket"
	"sync"
)

var onceToken sync.Once

// Unknown token reg
func (s *Service) unknownToken() {
	websocket.RegWSEndPoint(util.WSEndPoint)
	onceToken.Do(func() {
		if p, _ := rpc.GetSystemProperties(nil); p != nil {
			util.AddressType = util.IntToString(p.Ss58Format)
			util.BalanceAccuracy = util.IntToString(p.TokenDecimals)
		}
	})
}
