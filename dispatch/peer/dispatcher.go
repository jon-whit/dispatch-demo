// Package peer provides an implementation of the Dispatcher interface with
// support for remote peer dispatching which is based on dispatching calls
// using a consistent hashring.
package peer

import (
	"context"
	"fmt"

	"github.com/authzed/consistent"
	dispatchv1 "github.com/jon-whit/dispatch-demo/proto/dispatch/v1"
)

type PeerDispatcher struct {
	DispatchClient dispatchv1.DispatchServiceClient
}

func (p *PeerDispatcher) DispatchCheck(
	ctx context.Context,
	req *dispatchv1.DispatchCheckRequest,
) (*dispatchv1.DispatchCheckResponse, error) {
	fmt.Printf("calling peer.DispatchClient.DispatchCheck '%s:%s#%s' - this goes over the network\n", req.GetObjectType(), req.GetObjectId(), req.GetRelation())

	ctx = context.WithValue(ctx, consistent.CtxKey, []byte(req.GetObjectId()))
	return p.DispatchClient.DispatchCheck(ctx, req)
}
