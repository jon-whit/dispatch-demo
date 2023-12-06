package cache

import (
	"context"
	"fmt"

	"github.com/jon-whit/dispatch-demo/dispatch"
	dispatchv1 "github.com/jon-whit/dispatch-demo/proto/dispatch/v1"
)

type CachedDispatcher struct {
	Cache    map[string]*dispatchv1.DispatchCheckResponse
	Delegate dispatch.Dispatcher
}

// DispatchCheck implements dispatch.Dispatcher.
func (c *CachedDispatcher) DispatchCheck(
	ctx context.Context,
	req *dispatchv1.DispatchCheckRequest,
) (*dispatchv1.DispatchCheckResponse, error) {

	key := fmt.Sprintf("%s:%s#%s", req.GetObjectType(), req.GetObjectId(), req.GetRelation())

	fmt.Printf("cachedDispatcher.DispatchCheck (%s)\n", key)

	if resp, ok := c.Cache[key]; ok {
		fmt.Printf("serving '%s' from cache\n", key)
		return resp, nil
	} else {
		fmt.Printf("cache miss '%s'\n", key)
	}

	resp, err := c.Delegate.DispatchCheck(ctx, req)
	if err != nil {
		return nil, err
	}

	c.Cache[key] = resp
	return resp, nil
}

var _ dispatch.Dispatcher = (*CachedDispatcher)(nil)
