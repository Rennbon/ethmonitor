package main

import (
	"context"
	"github.com/Rennbon/ethmonitor"
	"log"
	"math/big"
	"time"
)

var _ ethmonitor.TxHandler = &Mock{}

type Mock struct {
}

// 持久化块高
func (m *Mock) SaveHeight(ctx context.Context, height *ethmonitor.BlockHeight) error {
	log.Println(height.String())
	return nil
}

// 启动monitor时初始化监听块高
func (m *Mock) LoadLastHeight(ctx context.Context) (*ethmonitor.BlockHeight, error) {
	return big.NewInt(1), nil
}

// 具体业务处理
func (m *Mock) Do(ctx context.Context, info *ethmonitor.TxInfo) {
	log.Println(info)
}

// 是否包含需要监控的合约地址
func (m *Mock) ContainContact(ctx context.Context, address ethmonitor.ContractAddress) bool {
	return true
}

// NOTE: 如果需要监听多个合约，可以将合约的abi合并成一个abiStr
func main() {
	opt := &ethmonitor.Options{
		RpcUrl: "http://localhost:8545",
		AbiStr: `
[
	{ "type" : "function", "name" : ""},
	{ "type" : "function", "name" : "balance", "stateMutability" : "view" },
	{ "type" : "function", "name" : "send", "inputs" : [ { "name" : "amount", "type" : "uint256" } ] },
	{ "type" : "function", "name" : "test", "inputs" : [ { "name" : "number", "type" : "uint32" } ] },
	{ "type" : "function", "name" : "string", "inputs" : [ { "name" : "inputs", "type" : "string" } ] },
	{ "type" : "function", "name" : "bool", "inputs" : [ { "name" : "inputs", "type" : "bool" } ] },
	{ "type" : "function", "name" : "address", "inputs" : [ { "name" : "inputs", "type" : "address" } ] },
	{ "type" : "function", "name" : "uint64[2]", "inputs" : [ { "name" : "inputs", "type" : "uint64[2]" } ] },
	{ "type" : "function", "name" : "uint64[]", "inputs" : [ { "name" : "inputs", "type" : "uint64[]" } ] },
	{ "type" : "function", "name" : "int8", "inputs" : [ { "name" : "inputs", "type" : "int8" } ] },
	{ "type" : "function", "name" : "bytes32", "inputs" : [ { "name" : "inputs", "type" : "bytes32" } ] },
	{ "type" : "function", "name" : "foo", "inputs" : [ { "name" : "inputs", "type" : "uint32" } ] },
	{ "type" : "function", "name" : "bar", "inputs" : [ { "name" : "inputs", "type" : "uint32" }, { "name" : "string", "type" : "uint16" } ] },
	{ "type" : "function", "name" : "slice", "inputs" : [ { "name" : "inputs", "type" : "uint32[2]" } ] },
	{ "type" : "function", "name" : "slice256", "inputs" : [ { "name" : "inputs", "type" : "uint256[2]" } ] },
	{ "type" : "function", "name" : "sliceAddress", "inputs" : [ { "name" : "inputs", "type" : "address[]" } ] },
	{ "type" : "function", "name" : "sliceMultiAddress", "inputs" : [ { "name" : "a", "type" : "address[]" }, { "name" : "b", "type" : "address[]" } ] },
	{ "type" : "function", "name" : "nestedArray", "inputs" : [ { "name" : "a", "type" : "uint256[2][2]" }, { "name" : "b", "type" : "address[]" } ] },
	{ "type" : "function", "name" : "nestedArray2", "inputs" : [ { "name" : "a", "type" : "uint8[][2]" } ] },
	{ "type" : "function", "name" : "nestedSlice", "inputs" : [ { "name" : "a", "type" : "uint8[][]" } ] },
	{ "type" : "function", "name" : "receive", "inputs" : [ { "name" : "memo", "type" : "bytes" }], "outputs" : [], "payable" : true, "stateMutability" : "payable" },
	{ "type" : "function", "name" : "fixedArrStr", "stateMutability" : "view", "inputs" : [ { "name" : "str", "type" : "string" }, { "name" : "fixedArr", "type" : "uint256[2]" } ] },
	{ "type" : "function", "name" : "fixedArrBytes", "stateMutability" : "view", "inputs" : [ { "name" : "bytes", "type" : "bytes" }, { "name" : "fixedArr", "type" : "uint256[2]" } ] },
	{ "type" : "function", "name" : "mixedArrStr", "stateMutability" : "view", "inputs" : [ { "name" : "str", "type" : "string" }, { "name" : "fixedArr", "type" : "uint256[2]" }, { "name" : "dynArr", "type" : "uint256[]" } ] },
	{ "type" : "function", "name" : "doubleFixedArrStr", "stateMutability" : "view", "inputs" : [ { "name" : "str", "type" : "string" }, { "name" : "fixedArr1", "type" : "uint256[2]" }, { "name" : "fixedArr2", "type" : "uint256[3]" } ] },
	{ "type" : "function", "name" : "multipleMixedArrStr", "stateMutability" : "view", "inputs" : [ { "name" : "str", "type" : "string" }, { "name" : "fixedArr1", "type" : "uint256[2]" }, { "name" : "dynArr", "type" : "uint256[]" }, { "name" : "fixedArr2", "type" : "uint256[3]" } ] },
	{ "type" : "function", "name" : "overloadedNames", "stateMutability" : "view", "inputs": [ { "components": [ { "internalType": "uint256", "name": "_f",	"type": "uint256" }, { "internalType": "uint256", "name": "__f", "type": "uint256"}, { "internalType": "uint256", "name": "f", "type": "uint256"}],"internalType": "struct Overloader.F", "name": "f","type": "tuple"}]}
]`,
		Handler: &Mock{},
	}
	monitor, err := ethmonitor.New(opt)
	if err != nil {
		panic(err)
	}
	monitor.Run()
	time.Sleep(time.Minute)
	monitor.Cancel()
}
