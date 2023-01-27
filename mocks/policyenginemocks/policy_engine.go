// Code generated by mockery v2.14.1. DO NOT EDIT.

package policyenginemocks

import (
	context "context"

	apitypes "github.com/hyperledger/firefly-transaction-manager/pkg/apitypes"

	ffcapi "github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"

	mock "github.com/stretchr/testify/mock"

	policyengine "github.com/hyperledger/firefly-transaction-manager/pkg/policyengine"
)

// PolicyEngine is an autogenerated mock type for the PolicyEngine type
type PolicyEngine struct {
	mock.Mock
}

// Execute provides a mock function with given fields: ctx, cAPI, mtx
func (_m *PolicyEngine) Execute(ctx context.Context, cAPI *policyengine.ToolkitAPI, mtx *apitypes.ManagedTX) (policyengine.UpdateType, ffcapi.ErrorReason, error) {
	ret := _m.Called(ctx, cAPI, mtx)

	var r0 policyengine.UpdateType
	if rf, ok := ret.Get(0).(func(context.Context, *policyengine.ToolkitAPI, *apitypes.ManagedTX) policyengine.UpdateType); ok {
		r0 = rf(ctx, cAPI, mtx)
	} else {
		r0 = ret.Get(0).(policyengine.UpdateType)
	}

	var r1 ffcapi.ErrorReason
	if rf, ok := ret.Get(1).(func(context.Context, *policyengine.ToolkitAPI, *apitypes.ManagedTX) ffcapi.ErrorReason); ok {
		r1 = rf(ctx, cAPI, mtx)
	} else {
		r1 = ret.Get(1).(ffcapi.ErrorReason)
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(context.Context, *policyengine.ToolkitAPI, *apitypes.ManagedTX) error); ok {
		r2 = rf(ctx, cAPI, mtx)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

type mockConstructorTestingTNewPolicyEngine interface {
	mock.TestingT
	Cleanup(func())
}

// NewPolicyEngine creates a new instance of PolicyEngine. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewPolicyEngine(t mockConstructorTestingTNewPolicyEngine) *PolicyEngine {
	mock := &PolicyEngine{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
