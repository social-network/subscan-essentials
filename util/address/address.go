package address

import (
	"github.com/social-network/netscan/util"
	"github.com/social-network/netscan/util/ss58"
)

func SS58Address(address string) string {
	return ss58.Encode(address, util.StringToInt(util.AddressType))
}
