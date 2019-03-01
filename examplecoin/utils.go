package examplecoin

import (
	"fmt"
	"github.com/algorand/go-algorand/daemon/algod/api/client"
	"github.com/algorand/go-algorand/daemon/algod/api/client/models"
	"github.com/algorand/go-algorand/protocol"
)

// BuildInitializeNote takes in the desired supply and produces a blob for your note field
func BuildInitializeNote(supply uint64) (initializeBlob client.BytesBase64) {
	initializeBlob = client.BytesBase64(protocol.Encode(NoteField{
		Type: NoteInitialize,
		Initialize: Initialize{
			Supply: supply,
		},
	}))
	return
}

// BuildInitializeNote takes in the desired recipient as well as amount to send, and produces a blob for your note field
func BuildTransferNote(amount uint64, to string) (transferBlob client.BytesBase64) {
	transferBlob = client.BytesBase64(protocol.Encode(NoteField{
		Type: NoteTransfer,
		Transfer: Transfer{
			Amount:      amount,
			Destination: to,
		},
	}))
	return
}

func processInitialize(curState map[string]uint64, initialize Initialize, wrappingTxn models.Transaction) (map[string]uint64, error) {
	if len(curState) != 0 {
		return curState, fmt.Errorf("attempted to process an initialize message against a ledger that was already initialized")
	}
	curState[wrappingTxn.From] = initialize.Supply
	return curState, nil
}

func processTransfer(curState map[string]uint64, transfer Transfer, wrappingTxn models.Transaction) (map[string]uint64, error) {
	if transfer.Source != wrappingTxn.From {
		return curState, fmt.Errorf("transaction submitted by %s tries to spend %s's examplecoin", wrappingTxn.From, transfer.Source)
	}
	senderBalance, exists := curState[transfer.Source]
	if !exists {
		return curState, fmt.Errorf("sender %v does not exist in the ledger", transfer.Source)
	}
	if transfer.Amount > senderBalance {
		return curState, fmt.Errorf("sender %v is trying to spend %d examplecoin, greater than balance %d", transfer.Source, transfer.Amount, senderBalance)
	}
	curState[transfer.Source] = senderBalance - transfer.Amount
	receiverBalance := curState[transfer.Destination]
	curState[transfer.Destination] = receiverBalance + transfer.Amount

	return curState, nil
}
