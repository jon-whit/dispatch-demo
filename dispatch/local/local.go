package local

import (
	"context"
	"fmt"

	"github.com/jon-whit/dispatch-demo/dispatch"
	dispatchv1 "github.com/jon-whit/dispatch-demo/proto/dispatch/v1"
)

type LocalDispatcher struct {
	Delegate dispatch.Dispatcher
}

var _ dispatch.Dispatcher = (*LocalDispatcher)(nil)

// DispatchCheck implements dispatch.Dispatcher.
func (l *LocalDispatcher) DispatchCheck(
	ctx context.Context,
	req *dispatchv1.DispatchCheckRequest,
) (*dispatchv1.DispatchCheckResponse, error) {

	if req.GetObjectType() == "document" && req.GetObjectId() == "1" && req.GetRelation() == "editor" {
		return &dispatchv1.DispatchCheckResponse{
			Allowed: true,
		}, nil
	}

	if req.GetObjectType() == "document" && req.GetObjectId() == "1" && req.GetRelation() == "viewer" {
		fmt.Println("(document:1#viewer) - dispatching document:1#editor")
		return l.Delegate.DispatchCheck(ctx, &dispatchv1.DispatchCheckRequest{
			ObjectType: req.GetObjectType(),
			ObjectId:   req.GetObjectId(),
			Relation:   "editor", // rewritten relation
		})
	}

	return nil, fmt.Errorf("unexpected evaluation path")
}
