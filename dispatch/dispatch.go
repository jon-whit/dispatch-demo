package dispatch

import (
	"context"

	dispatchv1 "github.com/jon-whit/dispatch-demo/proto/dispatch/v1"
)

type Dispatcher interface {
	DispatchCheck(context.Context, *dispatchv1.DispatchCheckRequest) (*dispatchv1.DispatchCheckResponse, error)
}
