// Code generated by mockery v2.26.1. DO NOT EDIT.

package persistencemocks

import (
	context "context"

	apitypes "github.com/hyperledger/firefly-transaction-manager/pkg/apitypes"

	ffcapi "github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"

	fftypes "github.com/hyperledger/firefly-common/pkg/fftypes"

	mock "github.com/stretchr/testify/mock"

	persistence "github.com/hyperledger/firefly-transaction-manager/internal/persistence"
)

// TransactionPersistence is an autogenerated mock type for the TransactionPersistence type
type TransactionPersistence struct {
	mock.Mock
}

// AddTransactionConfirmations provides a mock function with given fields: ctx, txID, clearExisting, confirmations
func (_m *TransactionPersistence) AddTransactionConfirmations(ctx context.Context, txID string, clearExisting bool, confirmations ...apitypes.BlockInfo) error {
	_va := make([]interface{}, len(confirmations))
	for _i := range confirmations {
		_va[_i] = confirmations[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, txID, clearExisting)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, bool, ...apitypes.BlockInfo) error); ok {
		r0 = rf(ctx, txID, clearExisting, confirmations...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteTransaction provides a mock function with given fields: ctx, txID
func (_m *TransactionPersistence) DeleteTransaction(ctx context.Context, txID string) error {
	ret := _m.Called(ctx, txID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, txID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetTransactionByID provides a mock function with given fields: ctx, txID
func (_m *TransactionPersistence) GetTransactionByID(ctx context.Context, txID string) (*apitypes.ManagedTX, error) {
	ret := _m.Called(ctx, txID)

	var r0 *apitypes.ManagedTX
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*apitypes.ManagedTX, error)); ok {
		return rf(ctx, txID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *apitypes.ManagedTX); ok {
		r0 = rf(ctx, txID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*apitypes.ManagedTX)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, txID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetTransactionByIDWithHistory provides a mock function with given fields: ctx, txID
func (_m *TransactionPersistence) GetTransactionByIDWithHistory(ctx context.Context, txID string) (*apitypes.TXWithStatus, error) {
	ret := _m.Called(ctx, txID)

	var r0 *apitypes.TXWithStatus
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*apitypes.TXWithStatus, error)); ok {
		return rf(ctx, txID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *apitypes.TXWithStatus); ok {
		r0 = rf(ctx, txID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*apitypes.TXWithStatus)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, txID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetTransactionByNonce provides a mock function with given fields: ctx, signer, nonce
func (_m *TransactionPersistence) GetTransactionByNonce(ctx context.Context, signer string, nonce *fftypes.FFBigInt) (*apitypes.ManagedTX, error) {
	ret := _m.Called(ctx, signer, nonce)

	var r0 *apitypes.ManagedTX
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *fftypes.FFBigInt) (*apitypes.ManagedTX, error)); ok {
		return rf(ctx, signer, nonce)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, *fftypes.FFBigInt) *apitypes.ManagedTX); ok {
		r0 = rf(ctx, signer, nonce)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*apitypes.ManagedTX)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, *fftypes.FFBigInt) error); ok {
		r1 = rf(ctx, signer, nonce)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetTransactionConfirmations provides a mock function with given fields: ctx, txID
func (_m *TransactionPersistence) GetTransactionConfirmations(ctx context.Context, txID string) ([]apitypes.BlockInfo, error) {
	ret := _m.Called(ctx, txID)

	var r0 []apitypes.BlockInfo
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) ([]apitypes.BlockInfo, error)); ok {
		return rf(ctx, txID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) []apitypes.BlockInfo); ok {
		r0 = rf(ctx, txID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]apitypes.BlockInfo)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, txID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetTransactionReceipt provides a mock function with given fields: ctx, txID
func (_m *TransactionPersistence) GetTransactionReceipt(ctx context.Context, txID string) (*ffcapi.TransactionReceiptResponse, error) {
	ret := _m.Called(ctx, txID)

	var r0 *ffcapi.TransactionReceiptResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*ffcapi.TransactionReceiptResponse, error)); ok {
		return rf(ctx, txID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *ffcapi.TransactionReceiptResponse); ok {
		r0 = rf(ctx, txID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*ffcapi.TransactionReceiptResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, txID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// InsertTransaction provides a mock function with given fields: ctx, tx
func (_m *TransactionPersistence) InsertTransaction(ctx context.Context, tx *apitypes.ManagedTX) error {
	ret := _m.Called(ctx, tx)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *apitypes.ManagedTX) error); ok {
		r0 = rf(ctx, tx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ListTransactionsByCreateTime provides a mock function with given fields: ctx, after, limit, dir
func (_m *TransactionPersistence) ListTransactionsByCreateTime(ctx context.Context, after *apitypes.ManagedTX, limit int, dir persistence.SortDirection) ([]*apitypes.ManagedTX, error) {
	ret := _m.Called(ctx, after, limit, dir)

	var r0 []*apitypes.ManagedTX
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *apitypes.ManagedTX, int, persistence.SortDirection) ([]*apitypes.ManagedTX, error)); ok {
		return rf(ctx, after, limit, dir)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *apitypes.ManagedTX, int, persistence.SortDirection) []*apitypes.ManagedTX); ok {
		r0 = rf(ctx, after, limit, dir)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*apitypes.ManagedTX)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *apitypes.ManagedTX, int, persistence.SortDirection) error); ok {
		r1 = rf(ctx, after, limit, dir)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListTransactionsByNonce provides a mock function with given fields: ctx, signer, after, limit, dir
func (_m *TransactionPersistence) ListTransactionsByNonce(ctx context.Context, signer string, after *fftypes.FFBigInt, limit int, dir persistence.SortDirection) ([]*apitypes.ManagedTX, error) {
	ret := _m.Called(ctx, signer, after, limit, dir)

	var r0 []*apitypes.ManagedTX
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *fftypes.FFBigInt, int, persistence.SortDirection) ([]*apitypes.ManagedTX, error)); ok {
		return rf(ctx, signer, after, limit, dir)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, *fftypes.FFBigInt, int, persistence.SortDirection) []*apitypes.ManagedTX); ok {
		r0 = rf(ctx, signer, after, limit, dir)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*apitypes.ManagedTX)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, *fftypes.FFBigInt, int, persistence.SortDirection) error); ok {
		r1 = rf(ctx, signer, after, limit, dir)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListTransactionsPending provides a mock function with given fields: ctx, afterSequenceID, limit, dir
func (_m *TransactionPersistence) ListTransactionsPending(ctx context.Context, afterSequenceID string, limit int, dir persistence.SortDirection) ([]*apitypes.ManagedTX, error) {
	ret := _m.Called(ctx, afterSequenceID, limit, dir)

	var r0 []*apitypes.ManagedTX
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, int, persistence.SortDirection) ([]*apitypes.ManagedTX, error)); ok {
		return rf(ctx, afterSequenceID, limit, dir)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, int, persistence.SortDirection) []*apitypes.ManagedTX); ok {
		r0 = rf(ctx, afterSequenceID, limit, dir)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*apitypes.ManagedTX)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, int, persistence.SortDirection) error); ok {
		r1 = rf(ctx, afterSequenceID, limit, dir)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SetTransactionReceipt provides a mock function with given fields: ctx, txID, receipt
func (_m *TransactionPersistence) SetTransactionReceipt(ctx context.Context, txID string, receipt *ffcapi.TransactionReceiptResponse) error {
	ret := _m.Called(ctx, txID, receipt)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *ffcapi.TransactionReceiptResponse) error); ok {
		r0 = rf(ctx, txID, receipt)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateTransaction provides a mock function with given fields: ctx, txID, updates
func (_m *TransactionPersistence) UpdateTransaction(ctx context.Context, txID string, updates *apitypes.TXUpdates) error {
	ret := _m.Called(ctx, txID, updates)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *apitypes.TXUpdates) error); ok {
		r0 = rf(ctx, txID, updates)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewTransactionPersistence interface {
	mock.TestingT
	Cleanup(func())
}

// NewTransactionPersistence creates a new instance of TransactionPersistence. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewTransactionPersistence(t mockConstructorTestingTNewTransactionPersistence) *TransactionPersistence {
	mock := &TransactionPersistence{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}