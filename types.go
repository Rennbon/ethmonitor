package ethmonitor

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
)

type Message = types.Message
type Receipt = types.Receipt
type BlockHeight = big.Int
type Address = common.Address
type Amount = big.Int
type ContractAddress = string
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
