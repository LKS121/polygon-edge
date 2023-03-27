package service

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/0xPolygon/polygon-edge/consensus/polybft/wallet"
	"github.com/0xPolygon/polygon-edge/types"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/umbracle/ethgo"
)

var aaInvokerAddress = types.StringToAddress("0x301")

func Test_AARelayerService_Start(t *testing.T) {
	t.Parallel()

	t.Run("executeJob_ok", func(t *testing.T) {
		t.Parallel()

		var log = []*ethgo.Log{
			{BlockNumber: 1, Topics: []ethgo.Hash{ethgo.ZeroHash}}, {BlockNumber: 5, Topics: []ethgo.Hash{ethgo.ZeroHash}}, {BlockNumber: 8, Topics: []ethgo.Hash{ethgo.ZeroHash}},
		}
		receipt := &ethgo.Receipt{GasUsed: 10, BlockHash: ethgo.ZeroHash, TransactionHash: ethgo.ZeroHash, Logs: log, Status: 1}
		pool := new(dummyAApool)

		account, err := wallet.GenerateAccount()
		require.NoError(t, err)

		state := new(dummyAATxState)
		state.On("Update", mock.Anything).Return(nil)

		aaTxSender := new(dummyAATxSender)
		aaTxSender.On("SendTransaction", mock.Anything, mock.Anything).
			Return(ethgo.ZeroHash, nil).Once()
		aaTxSender.On("WaitForReceipt", mock.Anything, mock.Anything, time.Second*3, 5).
			Return(receipt, nil)
		aaTxSender.On("GetNonce", mock.Anything).
			Return(uint64(0), error(nil)).Once()

		aaRelayerService, err := NewAARelayerService(aaTxSender, pool, state, account.Ecdsa, aaInvokerAddress, hclog.NewNullLogger(),
			WithPullTime(2*time.Second), WithReceiptDelay(time.Second*3), WithNumRetries(5))
		require.NoError(t, err)

		tx := getDummyTxs()[0]

		require.NoError(t, tx.Tx.MakeSignature(aaInvokerAddress, chainID, account.Ecdsa))
		require.NoError(t, aaRelayerService.executeJob(context.Background(), tx))
	})

	t.Run("executeJob_sendTransactionError", func(t *testing.T) {
		t.Parallel()

		pool := new(dummyAApool)

		account, err := wallet.GenerateAccount()
		require.NoError(t, err)

		state := new(dummyAATxState)
		state.On("Update", mock.Anything).Return(nil)

		aaTxSender := new(dummyAATxSender)
		aaTxSender.On("SendTransaction", mock.Anything, mock.Anything).
			Return(ethgo.ZeroHash, errors.New("not nil")).Once()
		aaTxSender.On("WaitForReceipt", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(&ethgo.Receipt{GasUsed: 10, BlockHash: ethgo.ZeroHash, TransactionHash: ethgo.ZeroHash}, nil).Once()
		aaTxSender.On("GetNonce", mock.Anything).
			Return(uint64(0), error(nil)).Once()

		aaRelayerService, err := NewAARelayerService(aaTxSender, pool, state, account.Ecdsa, aaInvokerAddress, hclog.NewNullLogger())
		require.NoError(t, err)

		tx := getDummyTxs()[1]

		require.NoError(t, tx.Tx.MakeSignature(aaInvokerAddress, chainID, account.Ecdsa))
		require.Error(t, aaRelayerService.executeJob(context.Background(), tx))
	})

	t.Run("executeJob_WaitForReceiptError", func(t *testing.T) {
		t.Parallel()

		account, err := wallet.GenerateAccount()
		require.NoError(t, err)

		pool := new(dummyAApool)

		state := new(dummyAATxState)
		state.On("Update", mock.Anything).Return(nil)

		aaTxSender := new(dummyAATxSender)
		aaTxSender.On("SendTransaction", mock.Anything, mock.Anything).
			Return(ethgo.ZeroHash, nil).Once()
		aaTxSender.On("WaitForReceipt", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(&ethgo.Receipt{GasUsed: 10, BlockHash: ethgo.ZeroHash, TransactionHash: ethgo.ZeroHash}, errors.New("not nil")).Once()
		aaTxSender.On("GetNonce", mock.Anything).
			Return(uint64(0), error(nil)).Once()

		aaRelayerService, err := NewAARelayerService(aaTxSender, pool, state, account.Ecdsa, aaInvokerAddress, hclog.NewNullLogger())
		require.NoError(t, err)

		tx := getDummyTxs()[2]

		require.NoError(t, tx.Tx.MakeSignature(aaInvokerAddress, chainID, account.Ecdsa))
		require.Error(t, aaRelayerService.executeJob(context.Background(), tx))
	})

	t.Run("executeJob_UpdateError", func(t *testing.T) {
		t.Parallel()

		pool := new(dummyAApool)

		account, err := wallet.GenerateAccount()
		require.NoError(t, err)

		state := new(dummyAATxState)
		state.On("Update", mock.Anything).Return(errors.New("not nil"))

		aaTxSender := new(dummyAATxSender)
		aaTxSender.On("SendTransaction", mock.Anything, mock.Anything).
			Return(ethgo.ZeroHash, errors.New("not nil")).Once()
		aaTxSender.On("WaitForReceipt", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(&ethgo.Receipt{GasUsed: 10, BlockHash: ethgo.ZeroHash, TransactionHash: ethgo.ZeroHash}, nil).Once()
		aaTxSender.On("GetNonce", mock.Anything).
			Return(uint64(0), error(nil)).Once()

		aaRelayerService, err := NewAARelayerService(aaTxSender, pool, state, account.Ecdsa, aaInvokerAddress, hclog.NewNullLogger())
		require.NoError(t, err)

		tx := getDummyTxs()[3]

		require.NoError(t, tx.Tx.MakeSignature(aaInvokerAddress, chainID, account.Ecdsa))
		require.Error(t, aaRelayerService.executeJob(context.Background(), tx))
	})

	t.Run("executeJob_NetError", func(t *testing.T) {
		t.Parallel()

		account, err := wallet.GenerateAccount()
		require.NoError(t, err)

		state := new(dummyAATxState)
		state.On("Update", mock.Anything).Return(errors.New("not nil"))

		pool := new(dummyAApool)
		pool.On("Push", mock.Anything)

		aaTxSender := new(dummyAATxSender)
		aaTxSender.On("SendTransaction", mock.Anything, mock.Anything).
			Return(ethgo.ZeroHash, net.ErrClosed).Once()
		aaTxSender.On("WaitForReceipt", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(&ethgo.Receipt{GasUsed: 10, BlockHash: ethgo.ZeroHash, TransactionHash: ethgo.ZeroHash}, nil).Once()
		aaTxSender.On("GetNonce", mock.Anything).
			Return(uint64(0), error(nil)).Once()

		aaRelayerService, err := NewAARelayerService(aaTxSender, pool, state, account.Ecdsa, aaInvokerAddress, hclog.NewNullLogger())
		require.NoError(t, err)

		tx := getDummyTxs()[4]

		require.NoError(t, tx.Tx.MakeSignature(aaInvokerAddress, chainID, account.Ecdsa))
		require.Error(t, aaRelayerService.executeJob(context.Background(), tx))
	})

	t.Run("executeJob_SecondUpdateError", func(t *testing.T) {
		t.Parallel()

		receipt := &ethgo.Receipt{GasUsed: 10, BlockHash: ethgo.ZeroHash, TransactionHash: ethgo.ZeroHash, Status: 1}

		account, err := wallet.GenerateAccount()
		require.NoError(t, err)

		state := new(dummyAATxState)
		state.On("Update", mock.Anything).Return(nil)
		state.On("Update", mock.Anything).Return(errors.New("not nil")).Once()

		pool := new(dummyAApool)
		pool.On("Push", mock.Anything)

		aaTxSender := new(dummyAATxSender)
		aaTxSender.On("SendTransaction", mock.Anything, mock.Anything).
			Return(ethgo.ZeroHash, nil).Once()
		aaTxSender.On("WaitForReceipt", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(receipt, nil).Once()
		aaTxSender.On("GetNonce", mock.Anything).Return(uint64(0), error(nil)).Once()

		tx := getDummyTxs()[0]

		aaRelayerService, err := NewAARelayerService(aaTxSender, pool, state, account.Ecdsa, aaInvokerAddress, hclog.NewNullLogger())
		require.NoError(t, err)

		require.NoError(t, tx.Tx.MakeSignature(aaInvokerAddress, chainID, account.Ecdsa))
		require.NoError(t, aaRelayerService.executeJob(context.Background(), tx))
	})

	t.Run("NewAARelayerService_errorGetNonce", func(t *testing.T) {
		t.Parallel()

		targetErr := errors.New("err")

		account, err := wallet.GenerateAccount()
		require.NoError(t, err)

		aaTxSender := new(dummyAATxSender)
		aaTxSender.On("GetNonce", mock.Anything).Return(uint64(0), targetErr).Once()

		_, err = NewAARelayerService(aaTxSender, nil, nil, account.Ecdsa, aaInvokerAddress, hclog.NewNullLogger())
		require.ErrorIs(t, err, targetErr)
	})
}

type dummyAApool struct {
	mock.Mock
}

func (p *dummyAApool) Push(stateTx *AAStateTransaction) {
	args := p.Called()
	_ = args
}

func (p *dummyAApool) Pop() *AAStateTransaction {
	args := p.Called()

	return args.Get(0).(*AAStateTransaction) //nolint:forcetypeassert
}

func (p *dummyAApool) Update(address types.Address) {
	args := p.Called(address)
	_ = args
}

func (p *dummyAApool) Init(txs []*AAStateTransaction) {
	args := p.Called(txs)

	_ = args
}
func (p *dummyAApool) Len() int {
	args := p.Called()

	return args.Int(0)
}

type dummyAATxState struct {
	mock.Mock
}

func (t *dummyAATxState) Add(transaction *AATransaction) (*AAStateTransaction, error) {
	args := t.Called()

	return args.Get(0).(*AAStateTransaction), args.Error(1) //nolint:forcetypeassert
}

func (t *dummyAATxState) Get(id string) (*AAStateTransaction, error) {
	args := t.Called(id)

	return args.Get(0).(*AAStateTransaction), args.Error(1) //nolint:forcetypeassert
}

func (t *dummyAATxState) GetAllPending() ([]*AAStateTransaction, error) {
	args := t.Called()

	return args.Get(0).([]*AAStateTransaction), args.Error(1) //nolint:forcetypeassert
}
func (t *dummyAATxState) GetAllQueued() ([]*AAStateTransaction, error) {
	args := t.Called()

	return args.Get(0).([]*AAStateTransaction), args.Error(1) //nolint:forcetypeassert
}
func (t *dummyAATxState) Update(stateTx *AAStateTransaction) error {
	args := t.Called()

	if stateTx.Status == StatusFailed {
		return errors.New("Update failed")
	}

	return args.Error(0)
}

type dummyAATxSender struct {
	mock.Mock

	test             *testing.T
	checkpointBlocks []uint64
}

func newDummyAATxSender(t *testing.T) *dummyAATxSender {
	t.Helper()

	return &dummyAATxSender{test: t}
}

func (d *dummyAATxSender) WaitForReceipt(
	ctx context.Context, hash ethgo.Hash, delay time.Duration, numRetries int) (*ethgo.Receipt, error) {
	args := d.Called(ctx, hash, delay, numRetries)

	return args.Get(0).(*ethgo.Receipt), args.Error(1) //nolint:forcetypeassert
}

func (d *dummyAATxSender) SendTransaction(txn *ethgo.Transaction, key ethgo.Key) (ethgo.Hash, error) {
	args := d.Called(txn, key)

	return args.Get(0).(ethgo.Hash), args.Error(1) //nolint:forcetypeassert
}

func (d *dummyAATxSender) GetNonce(address ethgo.Address) (uint64, error) {
	args := d.Called(address)

	return args.Get(0).(uint64), args.Error(1) //nolint:forcetypeassert
}

func (d *dummyAATxSender) GetAANonce(invokerAddress, address ethgo.Address) (uint64, error) {
	args := d.Called(invokerAddress, address)

	return args.Get(0).(uint64), args.Error(1) //nolint:forcetypeassert
}