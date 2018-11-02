// Code generated by counterfeiter. DO NOT EDIT.
package fakes

import (
	"policy-server/store"
	"policy-server/uaa_client"
	"sync"
)

type PolicyGuard struct {
	CheckAccessStub        func(policies []store.Policy, tokenData uaa_client.CheckTokenResponse) (bool, error)
	checkAccessMutex       sync.RWMutex
	checkAccessArgsForCall []struct {
		policies  []store.Policy
		tokenData uaa_client.CheckTokenResponse
	}
	checkAccessReturns struct {
		result1 bool
		result2 error
	}
	checkAccessReturnsOnCall map[int]struct {
		result1 bool
		result2 error
	}
	IsNetworkAdminStub        func(subjectToken uaa_client.CheckTokenResponse) bool
	isNetworkAdminMutex       sync.RWMutex
	isNetworkAdminArgsForCall []struct {
		subjectToken uaa_client.CheckTokenResponse
	}
	isNetworkAdminReturns struct {
		result1 bool
	}
	isNetworkAdminReturnsOnCall map[int]struct {
		result1 bool
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *PolicyGuard) CheckAccess(policies []store.Policy, tokenData uaa_client.CheckTokenResponse) (bool, error) {
	var policiesCopy []store.Policy
	if policies != nil {
		policiesCopy = make([]store.Policy, len(policies))
		copy(policiesCopy, policies)
	}
	fake.checkAccessMutex.Lock()
	ret, specificReturn := fake.checkAccessReturnsOnCall[len(fake.checkAccessArgsForCall)]
	fake.checkAccessArgsForCall = append(fake.checkAccessArgsForCall, struct {
		policies  []store.Policy
		tokenData uaa_client.CheckTokenResponse
	}{policiesCopy, tokenData})
	fake.recordInvocation("CheckAccess", []interface{}{policiesCopy, tokenData})
	fake.checkAccessMutex.Unlock()
	if fake.CheckAccessStub != nil {
		return fake.CheckAccessStub(policies, tokenData)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fake.checkAccessReturns.result1, fake.checkAccessReturns.result2
}

func (fake *PolicyGuard) CheckAccessCallCount() int {
	fake.checkAccessMutex.RLock()
	defer fake.checkAccessMutex.RUnlock()
	return len(fake.checkAccessArgsForCall)
}

func (fake *PolicyGuard) CheckAccessArgsForCall(i int) ([]store.Policy, uaa_client.CheckTokenResponse) {
	fake.checkAccessMutex.RLock()
	defer fake.checkAccessMutex.RUnlock()
	return fake.checkAccessArgsForCall[i].policies, fake.checkAccessArgsForCall[i].tokenData
}

func (fake *PolicyGuard) CheckAccessReturns(result1 bool, result2 error) {
	fake.CheckAccessStub = nil
	fake.checkAccessReturns = struct {
		result1 bool
		result2 error
	}{result1, result2}
}

func (fake *PolicyGuard) CheckAccessReturnsOnCall(i int, result1 bool, result2 error) {
	fake.CheckAccessStub = nil
	if fake.checkAccessReturnsOnCall == nil {
		fake.checkAccessReturnsOnCall = make(map[int]struct {
			result1 bool
			result2 error
		})
	}
	fake.checkAccessReturnsOnCall[i] = struct {
		result1 bool
		result2 error
	}{result1, result2}
}

func (fake *PolicyGuard) IsNetworkAdmin(subjectToken uaa_client.CheckTokenResponse) bool {
	fake.isNetworkAdminMutex.Lock()
	ret, specificReturn := fake.isNetworkAdminReturnsOnCall[len(fake.isNetworkAdminArgsForCall)]
	fake.isNetworkAdminArgsForCall = append(fake.isNetworkAdminArgsForCall, struct {
		subjectToken uaa_client.CheckTokenResponse
	}{subjectToken})
	fake.recordInvocation("IsNetworkAdmin", []interface{}{subjectToken})
	fake.isNetworkAdminMutex.Unlock()
	if fake.IsNetworkAdminStub != nil {
		return fake.IsNetworkAdminStub(subjectToken)
	}
	if specificReturn {
		return ret.result1
	}
	return fake.isNetworkAdminReturns.result1
}

func (fake *PolicyGuard) IsNetworkAdminCallCount() int {
	fake.isNetworkAdminMutex.RLock()
	defer fake.isNetworkAdminMutex.RUnlock()
	return len(fake.isNetworkAdminArgsForCall)
}

func (fake *PolicyGuard) IsNetworkAdminArgsForCall(i int) uaa_client.CheckTokenResponse {
	fake.isNetworkAdminMutex.RLock()
	defer fake.isNetworkAdminMutex.RUnlock()
	return fake.isNetworkAdminArgsForCall[i].subjectToken
}

func (fake *PolicyGuard) IsNetworkAdminReturns(result1 bool) {
	fake.IsNetworkAdminStub = nil
	fake.isNetworkAdminReturns = struct {
		result1 bool
	}{result1}
}

func (fake *PolicyGuard) IsNetworkAdminReturnsOnCall(i int, result1 bool) {
	fake.IsNetworkAdminStub = nil
	if fake.isNetworkAdminReturnsOnCall == nil {
		fake.isNetworkAdminReturnsOnCall = make(map[int]struct {
			result1 bool
		})
	}
	fake.isNetworkAdminReturnsOnCall[i] = struct {
		result1 bool
	}{result1}
}

func (fake *PolicyGuard) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.checkAccessMutex.RLock()
	defer fake.checkAccessMutex.RUnlock()
	fake.isNetworkAdminMutex.RLock()
	defer fake.isNetworkAdminMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *PolicyGuard) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}
