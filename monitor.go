package ethmonitor

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
	"math/big"
	"sync"
	"time"
)

type Monitor interface {
	Run()
	Cancel()
}

var FieldTag = "monitor"

type Options struct {
	RpcUrl  string    // eth节点url
	AbiStr  string    // 组合abi,可以监控多个智能合约
	Handler TxHandler // 业务处理handler
	Logger  logrus.FieldLogger
}

type monitor struct {
	ctx     context.Context
	cancel  context.CancelFunc
	cli     *ethclient.Client
	hScan   HeightScanner
	decoder *abiDecoder
	handler TxHandler
	logger  logrus.FieldLogger
	sync.RWMutex
}

// 基本参数验证
func (opt *Options) check() error {
	if opt == nil {
		return errors.New("options nil reference")
	}
	if opt.RpcUrl == "" || opt.AbiStr == "" {
		return errors.New("rpcUrl and abiStr can't be empty")
	}
	if opt.Handler == nil {
		return errors.New("handler nil reference")
	}
	if opt.Logger == nil {
		opt.Logger = logrus.New()
	}
	return nil
}

// New 初始化eth 监控器
func New(opt *Options) (Monitor, error) {
	if err := opt.check(); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	cli, err := ethclient.DialContext(ctx, opt.RpcUrl)
	if err != nil {
		cancel()
		return nil, err
	}
	m := &monitor{
		ctx:     ctx,
		cancel:  cancel,
		cli:     cli,
		decoder: newAbiDecoder(opt.AbiStr),
		hScan:   opt.Handler,
		handler: opt.Handler,
		logger:  opt.Logger,
	}
	return m, nil
}

// HeightScanner 高度扫描器
type HeightScanner interface {
	// SaveHeight 持久化最新块高
	SaveHeight(ctx context.Context, height *BlockHeight) error
	// LoadLastHeight 加载上一次块高
	LoadLastHeight(ctx context.Context) (*BlockHeight, error)
}

// TxHandler 业务tx句柄
type TxHandler interface {
	HeightScanner
	// Do 处理命中的tx
	Do(ctx context.Context, info *TxInfo)
	// ContainContact 是否包含指定合约token
	// NOTE: 如果不满足也放行会在decode中抛出error "illegal tx"
	// 如果是多智能合约监听，可以使用map维护多个
	// 配套的，需要把这些合约的abi合并在初始化monitor时赋值给AbiStr，注意去重
	ContainContact(ctx context.Context, address ContractAddress) bool
}

func (m *monitor) Run() {
	lastBlockHeight, err := m.hScan.LoadLastHeight(m.ctx)
	if err != nil {
		panic(err)
	}
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			curIndex, _, err := m.getBlockHeight()
			if err != nil {
				m.logger.WithField(FieldTag, "getBlockHeight").Error(err)
				continue
			}
			if curIndex.Cmp(lastBlockHeight) == 0 {
				continue
			}
			if lastBlockHeight.Cmp(big.NewInt(0)) == 0 {
				lastBlockHeight.Set(curIndex)
				continue
			}
			start := big.NewInt(0).Set(lastBlockHeight)
			end := big.NewInt(0).Set(curIndex)
			m.blockListen(start, end)
			lastBlockHeight.Set(curIndex)
			err = m.hScan.SaveHeight(m.ctx, curIndex)
			if err != nil {
				m.logger.WithField(FieldTag, "saveHeight").Error(err)
			}
		case <-m.ctx.Done():
			m.logger.WithField(FieldTag, "close").Info()
			return
		}
	}
}

func (m *monitor) Cancel() {
	m.cancel()
}

func (m *monitor) getBlockHeight() (cur, highest *BlockHeight, err error) {
	sync, err := m.cli.SyncProgress(m.ctx)
	if err != nil {
		return nil, nil, err
	}
	if sync == nil {
		block, err := m.cli.HeaderByNumber(context.Background(), nil)
		if err != nil {
			return nil, nil, err
		}
		return block.Number, block.Number, nil
	} else {
		return big.NewInt(0).SetUint64(sync.CurrentBlock), big.NewInt(0).SetUint64(sync.HighestBlock), nil
	}
}

func (m *monitor) blockListen(start, end *BlockHeight) {
	for i := big.NewInt(0).Set(start); i.Cmp(end) < 0; i.Add(i, big.NewInt(1)) {
		var (
			block *types.Block
			err   error
		)
		for { // 失败阻塞，等待节点修复
			block, err = m.cli.BlockByNumber(context.Background(), i)
			if err != nil {
				m.logger.WithField(FieldTag, "blockByNumber").Errorf("height:%v error:%s", i.String(), err)
				time.Sleep(time.Second)
				continue
			}
			break
		}
		if block.Transactions().Len() > 0 {
			m.analyzeBlock(block)
		}

	}
}

func (m *monitor) analyzeBlock(block *types.Block) {
	for _, v := range block.Transactions() {
		signer := types.NewLondonSigner(v.ChainId())
		msg, err := v.AsMessage(signer, nil)
		if err != nil {
			m.logger.WithField(FieldTag, "asMessage").Error(err)
			continue
		}
		if msg.To() == nil {
			continue
		}
		if m.handler.ContainContact(m.ctx, msg.To().Hex()) {
			txInfo, err := m.analyzeTx(v.Hash(), &msg)
			if err != nil {
				m.logger.WithField(FieldTag, "analyzeTx").Error(err)
				continue
			}
			if txInfo != nil {
				m.handler.Do(m.ctx, txInfo)
			}
		}
	}
}

func (m *monitor) analyzeTx(txHash common.Hash, msg *Message) (*TxInfo, error) {
	defer func() {
		if err := recover(); err != nil {
			m.logger.WithField(FieldTag, "analyzeTx").Errorf("panic cover err:%v", err)
		}
	}()
	txInfo, isPending, err := m.cli.TransactionByHash(context.Background(), txHash)
	if err != nil {
		return nil, err
	}
	if isPending {
		return nil, nil
	}
	txRe, errTxRe := m.cli.TransactionReceipt(context.Background(), txHash)
	fee := big.NewInt(0)
	if errTxRe == nil && txRe != nil {
		fee = fee.SetUint64(txRe.GasUsed)
		fee = fee.Mul(fee, txInfo.GasPrice())
	}
	act, err := m.decoder.DecodeTxData(msg.Data())
	if err != nil {
		return nil, err
	}
	ti := &TxInfo{
		Message: msg,
		Receipt: txRe,
		Action:  act,
		TxHash:  txHash.Hex(),
		Fee:     fee,
		Height:  txRe.BlockNumber,
		Status:  txRe.Status == 1,
	}
	return ti, nil
}
