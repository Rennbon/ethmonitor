package ethmonitor

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
)

type Message = types.Message // 消息体
type Receipt = types.Receipt // 消息回执，验证有效性等
type BlockHeight = big.Int   // 块高
type Address = common.Address
type Amount = big.Int
type ContractAddress = string // 合约地址
type Singer types.Signer
type TxInfo struct {
	Message *Message     // 消息体
	Receipt *Receipt     // 回执
	Action  *Action      // 合约方法及参数
	TxHash  string       // 消息hash
	Height  *BlockHeight // 块高
	Fee     *Amount      // gas消耗
	Status  bool         // tx是否有效
}
