package ethmonitor

import (
	"errors"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"strings"
)

// Action 智能合约的方法
type Action struct {
	Method string                 // 合约方法
	Inputs map[string]interface{} //合约入参及对应的value
}

type abiDecoder struct {
	abi abi.ABI
}

func newAbiDecoder(abiStr string) *abiDecoder {
	a, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		panic(err)
	}
	return &abiDecoder{a}
}

func (d *abiDecoder) DecodeTxData(txData []byte) (*Action, error) {
	if len(txData) < 4 {
		return nil, errors.New("illegal tx")
	}
	method, err := d.abi.MethodById(txData[0:4])
	if err != nil {
		return nil, err
	}
	inputsMap := map[string]interface{}{}
	err = method.Inputs.UnpackIntoMap(inputsMap, txData[4:])
	if err != nil {
		return nil, err
	}
	act := &Action{
		Method: method.Name,
		Inputs: inputsMap,
	}
	return act, nil
}
