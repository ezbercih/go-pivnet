// This file was generated by counterfeiter
package commandsfakes

import (
	"sync"

	"github.com/pivotal-cf/go-pivnet/cmd/pivnet/commands"
)

type FakeCurlClient struct {
	MakeRequestStub        func(method string, body string, args []string) error
	makeRequestMutex       sync.RWMutex
	makeRequestArgsForCall []struct {
		method string
		body   string
		args   []string
	}
	makeRequestReturns struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeCurlClient) MakeRequest(method string, body string, args []string) error {
	var argsCopy []string
	if args != nil {
		argsCopy = make([]string, len(args))
		copy(argsCopy, args)
	}
	fake.makeRequestMutex.Lock()
	fake.makeRequestArgsForCall = append(fake.makeRequestArgsForCall, struct {
		method string
		body   string
		args   []string
	}{method, body, argsCopy})
	fake.recordInvocation("MakeRequest", []interface{}{method, body, argsCopy})
	fake.makeRequestMutex.Unlock()
	if fake.MakeRequestStub != nil {
		return fake.MakeRequestStub(method, body, args)
	} else {
		return fake.makeRequestReturns.result1
	}
}

func (fake *FakeCurlClient) MakeRequestCallCount() int {
	fake.makeRequestMutex.RLock()
	defer fake.makeRequestMutex.RUnlock()
	return len(fake.makeRequestArgsForCall)
}

func (fake *FakeCurlClient) MakeRequestArgsForCall(i int) (string, string, []string) {
	fake.makeRequestMutex.RLock()
	defer fake.makeRequestMutex.RUnlock()
	return fake.makeRequestArgsForCall[i].method, fake.makeRequestArgsForCall[i].body, fake.makeRequestArgsForCall[i].args
}

func (fake *FakeCurlClient) MakeRequestReturns(result1 error) {
	fake.MakeRequestStub = nil
	fake.makeRequestReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeCurlClient) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.makeRequestMutex.RLock()
	defer fake.makeRequestMutex.RUnlock()
	return fake.invocations
}

func (fake *FakeCurlClient) recordInvocation(key string, args []interface{}) {
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

var _ commands.CurlClient = new(FakeCurlClient)