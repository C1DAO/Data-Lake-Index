package bridge

import (
	"context"
	"errors"
	"math/big"

	"github.com/VMETA3/vmeta3-chain-indexer/bindings/l2bridge"
	"github.com/VMETA3/vmeta3-chain-indexer/db"

	"github.com/VMETA3/vmeta3-l2geth/accounts/abi/bind"
	"github.com/VMETA3/vmeta3-l2geth/common"
)

type WithdrawalsMap map[common.Hash][]db.Withdrawal

type Bridge interface {
	Address() common.Address
	GetWithdrawalsByBlockRange(uint64, uint64) (WithdrawalsMap, error)
	String() string
}

type implConfig struct {
	name string
	impl string
	addr string
}

var defaultBridgeCfgs = []*implConfig{
	{"Standard", "StandardBridge", L2StandardBridgeAddr},
}

var customBridgeCfgs = map[uint64][]*implConfig{
	// Mainnet
	10: {
		{"BitBTC", StandardBridgeImpl, "0x158F513096923fF2d3aab2BcF4478536de6725e2"},
		//{"DAI", "DAIBridge", "0x467194771dAe2967Aef3ECbEDD3Bf9a310C76C65"},
	},
	// Kovan
	69: {
		{"BitBTC", StandardBridgeImpl, "0x0CFb46528a7002a7D8877a5F7a69b9AaF1A9058e"},
		{"USX", StandardBridgeImpl, "0xB4d37826b14Cd3CB7257A2A5094507d701fe715f"},
		//{"DAI", "	DAIBridge", "0x467194771dAe2967Aef3ECbEDD3Bf9a310C76C65"},
	},
}

func BridgesByChainID(chainID *big.Int, client bind.ContractFilterer, ctx context.Context) (map[string]Bridge, error) {
	allCfgs := make([]*implConfig, 0)
	allCfgs = append(allCfgs, defaultBridgeCfgs...)
	allCfgs = append(allCfgs, customBridgeCfgs[chainID.Uint64()]...)

	bridges := make(map[string]Bridge)
	for _, bridge := range allCfgs {
		switch bridge.impl {
		case "StandardBridge":
			l2StandardBridgeAddress := common.HexToAddress(bridge.addr)
			l2StandardBridgeFilter, err := l2bridge.NewL2StandardBridgeFilterer(l2StandardBridgeAddress, client)
			if err != nil {
				return nil, err
			}

			standardBridge := &StandardBridge{
				name:     bridge.name,
				ctx:      ctx,
				address:  l2StandardBridgeAddress,
				client:   client,
				filterer: l2StandardBridgeFilter,
			}
			bridges[bridge.name] = standardBridge
		default:
			return nil, errors.New("unsupported bridge")
		}
	}
	return bridges, nil
}
