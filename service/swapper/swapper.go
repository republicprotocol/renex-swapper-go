package swapper

import (
	swapDomain "github.com/republicprotocol/renex-swapper-go/domain/swap"
)

type swapper struct {
	Adapter
	MutexCache
}

type Swapper interface {
}

func New(adapter Adapter) Swapper {
	return &swapper{
		adapter,
		NewMutexCache(),
	}
}

func (swapper *swapper) Run(reqCh <-chan swapDomain.Request, doneCh chan<- [32]byte, errCh chan<- error) {
	for {
		req, ok := <-reqCh
		if !ok {
			return
		}
		go func() {
			if !swapper.Lock(req.UID) {
				return
			}
			defer swapper.Unlock(req.UID)
			binder, err := swapper.BuildBinder(req)
			if err != nil {
				errCh <- err
				return
			}
			newSwap(binder).execute(req, errCh)
			doneCh <- req.UID
		}()
	}
}
