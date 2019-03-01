package examplecoin

import (
	"flag"
	"fmt"
	"github.com/algorand/go-algorand/daemon/algod/api/client"
	"github.com/algorand/go-algorand/data/basics"
	"github.com/algorand/go-algorand/protocol"
	"net/url"
	"os"
)

// This util is a tool that crawls the blockchain
// and outputs a csv file of the current examplecoin state.
var coinKey = flag.String("coinKey", "", "The pubkey of the coin's manager.")
var verboseFlag = flag.Bool("verbose", false, "Print extra debug info during operation.")

func main() {

	// these could be made into flag arguments,
	// or maybe you could read these in through a config file.
	// For this example, we're just going to hardcode them.
	localNodeURL := "fill me in!"
	algodToken := "fill me in!"

	algodURL, err := url.Parse(localNodeURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot parse algod URL %s: %v\n", localNodeURL, err)
		os.Exit(1)
	}

	// TODO evan decide if checksumAddress is appropriate or if we just want to pass strings around
	// consider also what is/is not in the golang SDK
	// see also client encoding.go:UnmarshalText instead of this
	masterKey, err := basics.UnmarshalChecksumAddress(*coinKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot unmarshal address from provided key %s: %v\n", *coinKey, err)
		os.Exit(1)
	}
	restClient := client.MakeRestClient(*algodURL, algodToken)
	masterKeyChecksumAddr := basics.Address(masterKey).GetChecksumAddress()

	results := make(map[client.ChecksumAddress]uint64)
	curRound := uint64(1)     // TODO evan make this a flag
	finalRound := uint64(305) // TODO evan make this a flag
	sawInitializeMessage := false
	for {
		if curRound > finalRound {
			break
		}

		if *verboseFlag {
			fmt.Printf("Checking round %d..\n", curRound)
		}

		txns, err := restClient.TransactionsByAddr(masterKeyChecksumAddr.String(), curRound, curRound)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot fetch transactions from %d: %v\n", curRound, err)
			os.Exit(1)
		}

		for _, txn := range txns.Transactions {
			if txn.ConfirmedRound != curRound {
				fmt.Fprintf(os.Stderr, "Confirmed round mismatch: found a txn claiming to be confirmed in round %d, in the block for round %d\n", txn.ConfirmedRound, curRound)
				os.Exit(1)
			}

			dec := protocol.NewDecoderBytes(txn.Note)

			for {
				var note NoteField
				err = dec.Decode(&note)
				if err != nil {
					break
				}

				switch note.Type {
				case NoteInitialize:
					if results, err = processInitialize(results, note.Initialize); err == nil {
						sawInitializeMessage = true
					} else {
						// TODO ejr bad path
					}
				case NoteTransfer:
					if results, err = processTransfer(results, note.Transfer); err == nil {
						// TODO ejr good path
					} else {
						// TODO ejr bad path
					}
				default:
					continue
				}
			}
		}
		curRound++
	}

	if !sawInitializeMessage {
		fmt.Fprintf(os.Stderr, "Did not find an examplecoin initialization from key %s\n", coinKey)
		os.Exit(1)
	}

	fmt.Printf("Collected balances for %d individuals\n", len(results))

	outfileName := "results.csv"
	outfile, err := os.Create(outfileName)
	if err != nil {
		fmt.Printf("Cannot create file %s: %v\n", outfileName, err)
		os.Exit(1)
	}
	defer outfile.Close()
	for user, balance := range results {
		output := fmt.Sprintf("%s,%d\n", user, balance)
		_, err = outfile.WriteString(output)
		if err != nil {
			fmt.Printf("Cannot write string \"%s\" to %s: %v\n", output, outfileName, err)
			os.Exit(1)
		}
	}
	fmt.Printf("Wrote results into %s\n", outfileName)
}
