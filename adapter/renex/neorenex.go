package renex

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/republicprotocol/renex-swapper-go/domain/swap"
)

type RenExOrder struct {
	ID     [32]byte
	Expiry int64
}

func Run(orderCh <-chan RenExOrder, reqCh chan<- swap.Request, errCh chan<- error) {
	for {
		order, ok := <-orderCh
		if !ok {
			return
		}
		go func(order RenExOrder) {
			if time.Now().Unix() > order.Expiry {
				return
			}
			req, err := buildRequest(order)
			if err != nil {
				errCh <- err
				return
			}
			reqCh <- req
		}(order)
	}
}

func (adapter *renexAdapter) buildRequest(order RenExOrder) (swap.Request, error) {
	logger := adapter.LoggerBuilder.New(order.ID)
	logger.LogInfo("building swap request")
	req := swap.Request{}
	req.UID = order.ID
	logger.LogInfo("waiting for the order match to be found")
	match, err := adapter.GetOrderMatch(order.ID, order.Expiry)
	if err != nil {
		return req, err
	}
	logger.LogInfo(fmt.Sprintf("matched with (%s%s%s)", pickColor(logger.UID), base64.StdEncoding.EncodeToString(match.ForeignOrderID[:], white)))
	sendToAddress, receiveFromAddress := adapter.GetAddresses(match.ReceiveToken, match.SendToken)
	req.SendToken = match.SendToken
	req.ReceiveToken = match.ReceiveToken
	req.SendValue = match.SendValue
	req.ReceiveValue = match.ReceiveValue
	if req.SendToken > req.ReceiveToken {
		req.GoesFirst = true
		rand.Read(req.Secret[:])
		req.SecretHash = sha256.Sum256(req.Secret[:])
		req.TimeLock = time.Now().Unix() + 48*60*60
	} else {
		req.GoesFirst = false
	}
	logger.LogInfo("communicating with matching trader")
	if err := adapter.SendSwapDetails(req.UID, SwapDetails{
		SecretHash:         req.SecretHash,
		TimeLock:           req.TimeLock,
		SendToAddress:      sendToAddress,
		ReceiveFromAddress: receiveFromAddress,
	}); err != nil {
		return req, err
	}
	foreignDetails, err := adapter.ReceiveSwapDetails(ordMatch.ForeignOrderID, timeStamp+48*60*60)
	if err != nil {
		return req, err
	}
	logger.LogInfo("communication successful")
	req.SendToAddress = foreignDetails.SendToAddress
	req.ReceiveFromAddress = foreignDetails.ReceiveFromAddress
	if !req.GoesFirst {
		req.SecretHash = foreignDetails.SecretHash
		req.TimeLock = foreignDetails.TimeLock
	}
	renex.PrintSwapRequest(req)
	return req, nil
}
