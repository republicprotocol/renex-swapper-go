package swapper

import (
	"fmt"
	"time"

	co "github.com/republicprotocol/co-go"
	swapDomain "github.com/republicprotocol/renex-swapper-go/domain/swap"
)

var ErrAlreadyInitiated = fmt.Errorf("")
var ErrRedeemedOrRefunded = fmt.Errorf("")

type swap struct {
	adapter BlockchainBinder
}

func newSwap(adapter BlockchainBinder) *swap {
	return &swap{
		adapter,
	}
}

func (swap *swap) execute(req swapDomain.Request, errCh chan<- error) {
	notifyToInitiateCh := make(chan struct{}, 1)
	notifyToRedeemOrRefundCh := make(chan struct{}, 1)
	doneCh := make(chan struct{}, 1)
	secretCh := make(chan [32]byte, 1)
	defer close(notifyToInitiateCh)
	defer close(notifyToRedeemOrRefundCh)
	defer close(doneCh)
	defer close(secretCh)
	if req.GoesFirst {
		co.ParBegin(
			func() {
				notifyToInitiateCh <- struct{}{}
				secretCh <- req.Secret
			},
			func() {
				swap.initiate(notifyToInitiateCh, notifyToRedeemOrRefundCh, errCh)
			},
			func() {
				swap.audit(notifyToRedeemOrRefundCh, errCh)
			},
			func() {
				swap.redeem(secretCh, notifyToRedeemOrRefundCh, doneCh, errCh)
			},
			func() {
				swap.refund(notifyToRedeemOrRefundCh, doneCh, errCh)
			},
		)
		return
	}
	co.ParBegin(
		func() {
			swap.audit(notifyToInitiateCh, errCh)
		},
		func() {
			swap.initiate(notifyToInitiateCh, notifyToRedeemOrRefundCh, errCh)
		},
		func() {
			swap.auditSecret(secretCh, errCh)
		},
		func() {
			swap.redeem(secretCh, notifyToRedeemOrRefundCh, doneCh, errCh)
		},
		func() {
			swap.refund(notifyToRedeemOrRefundCh, doneCh, errCh)
		},
	)
}

func (swap *swap) initiate(notifyCh <-chan struct{}, notifyNextCh chan<- struct{}, errCh chan<- error) {
	_, ok := <-notifyCh
	if !ok {
		return
	}
	for {
		if err := swap.adapter.Initiate(); err != nil {
			if err == ErrAlreadyInitiated {
				break
			}
			errCh <- err
			time.Sleep(1 * time.Minute)
			continue
		}
		break
	}
	notifyNextCh <- struct{}{}
	return
}

func (swap *swap) audit(notifyNextCh chan<- struct{}, errCh chan<- error) {
	for {
		if err := swap.adapter.Audit(); err != nil {
			errCh <- err
			time.Sleep(1 * time.Minute)
			continue
		}
		notifyNextCh <- struct{}{}
		return
	}
}

func (swap *swap) redeem(secretCh <-chan [32]byte, notifyCh <-chan struct{}, doneCh chan struct{}, errCh chan<- error) {
	secret, ok := <-secretCh
	if !ok {
		return
	}
	_, ok = <-notifyCh
	if !ok {
		return
	}
	for {
		select {
		case _, ok := <-doneCh:
			if !ok {
				return
			}
			return
		default:
			if err := swap.adapter.Redeem(secret); err != nil {
				if err == ErrRedeemedOrRefunded {
					doneCh <- struct{}{}
					return
				}
				errCh <- err
				time.Sleep(1 * time.Minute)
				continue
			}
			doneCh <- struct{}{}
			return
		}
	}
}

func (swap *swap) auditSecret(secretCh chan<- [32]byte, errCh chan<- error) {
	for {
		secret, err := swap.adapter.AuditSecret()
		if err != nil {
			errCh <- err
			time.Sleep(1 * time.Minute)
			continue
		}
		secretCh <- secret
		return
	}
}

func (swap *swap) refund(notifyCh <-chan struct{}, doneCh chan struct{}, errCh chan<- error) {
	_, ok := <-notifyCh
	if !ok {
		return
	}
	for {
		select {
		case _, ok := <-doneCh:
			if !ok {
				return
			}
			return
		default:
			if err := swap.adapter.Refund(); err != nil {
				if err == ErrRedeemedOrRefunded {
					doneCh <- struct{}{}
					return
				}
				errCh <- err
				time.Sleep(1 * time.Minute)
				continue
			}
			doneCh <- struct{}{}
			return
		}
	}
}
