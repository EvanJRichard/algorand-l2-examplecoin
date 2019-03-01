package examplecoin

import (
	"github.com/algorand/go-algorand/daemon/algod/api/client"
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
func BuildTransferNote(amount uint64, to client.ChecksumAddress) (transferBlob client.BytesBase64) {
	transferBlob = client.BytesBase64(protocol.Encode(NoteField{
		Type: NoteTransfer,
		Transfer: Transfer{
			Amount:      amount,
			Destination: to,
		},
	}))
	return
}

func processInitialize(curState map[client.ChecksumAddress]uint64, initialize Initialize) (map[client.ChecksumAddress]uint64, error) {
	return curState, nil
}

func processTransfer(curState map[client.ChecksumAddress]uint64, transfer Transfer) (map[client.ChecksumAddress]uint64, error) {
	return curState, nil
}
