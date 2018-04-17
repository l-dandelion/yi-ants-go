package module

import (
	"net"
	"strconv"

	"github.com/l-dandelion/yi-ants-go/lib/constant"
)

/*
 * implementation of interface net.Addr
 */
type Addr struct {
	network string //net protocol
	address string //net address
}

/*
 * get the net protocol
 */
func (addr *Addr) Network() string {
	return addr.network
}

/*
 * get the net address
 */
func (addr *Addr) String() string {
	return addr.address
}

/*
 * if success: new and return a net.Addr according to network, ip and port
 * else return an error
 */
func NewAddr(network string, ip string, port uint64) (net.Addr, *constant.YiError) {
	if network != "http" && network != "https" {
		return nil, constant.NewYiErrorf(constant.ERR_NEW_ADDRESS,
			"Illegal network for module address: %s", network)
	}

	if parsedIP := net.ParseIP(ip); parsedIP == nil {
		return nil, constant.NewYiErrorf(constant.ERR_NEW_ADDRESS,
			"Illegal IP for module address: %s", ip)
	}

	return &Addr{
		network: network,
		address: ip + ":" + strconv.Itoa(int(port)),
	}, nil
}
