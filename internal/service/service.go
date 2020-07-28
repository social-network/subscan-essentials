package service

import (
	"fmt"
	"github.com/go-kratos/kratos/pkg/log"
	"github.com/social-network/scale.go/source"
	"github.com/social-network/scale.go/types"
	"github.com/social-network/netscan/internal/dao"
	"github.com/social-network/netscan/plugins"
	"github.com/social-network/netscan/util"
	"github.com/social-network/substrate-api-rpc/metadata"
	"github.com/social-network/substrate-api-rpc/websocket"
	"io/ioutil"
	"strings"
)

// Service
type Service struct {
	dao dao.IDao
}

// New new a service and return.
func New() (s *Service) {
	d := dao.New()
	s = &Service{
		dao: d,
	}
	websocket.RegWSEndPoint(util.WSEndPoint)
	s.initSubRuntimeLatest()

	for _, plugin := range plugins.RegisteredPlugins {
		plugin.InitDao(d)
	}
	return s
}

// Close close the resource.
func (s *Service) Close() {
	s.dao.Close()
}

func (s *Service) initSubRuntimeLatest() {
	// reg network custom type
	defer func() {
		go s.unknownToken()
		if c, err := readTypeRegistry(); err == nil {
			types.RegCustomTypes(source.LoadTypeRegistry(c))
			if unknown := metadata.Decoder.CheckRegistry(); len(unknown) > 0 {
				log.Warn("Found unknown type %s", strings.Join(unknown, ", "))
			}
		}
	}()

	// find db
	if recent := s.dao.RuntimeVersionRecent(); recent != nil && strings.HasPrefix(recent.RawData, "0x") {
		metadata.Latest(&metadata.RuntimeRaw{Spec: recent.SpecVersion, Raw: recent.RawData})
		return
	}
	// find metadata for blockChain
	if raw := s.regCodecMetadata(); strings.HasPrefix(raw, "0x") {
		metadata.Latest(&metadata.RuntimeRaw{Spec: 1, Raw: raw})
		return
	}
	panic("Can not find chain metadata, please check network")
}

// read custom registry from local or remote
func readTypeRegistry() ([]byte, error) {
	return ioutil.ReadFile(fmt.Sprintf("../configs/source/%s.json", util.NetworkNode))
}
