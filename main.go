package main

import (
	"flag"
	"fmt"
	"github.com/algorand/algorand-l2-examplecoin/examplecoin"
	"github.com/algorand/go-algorand/daemon/algod/api/client"
	"github.com/algorand/go-algorand/protocol"
	"net/url"
	"os"
)

// This util is a tool that crawls the blockchain
// and outputs a csv file of the current examplecoin state.
var coinKey = flag.String("coinkey", "", "The pubkey of the coin's manager.")
var verboseFlag = flag.Bool("verbose", false, "Print extra debug info during operation.")
var firstRound = flag.Uint64("firstround", 1, "the first round from which to start scanning")
var lastRound = flag.Uint64("lastround", 301, "the last round at which to stop scanning")

func main() {
	flag.Parse()
	// These could be made into flag arguments like coinKey,
	// or maybe you could read these in through a config file.
	// For this example, we're just going to hardcode them.
	localNodeURL := "http://127.0.0.1:51275"                                         // TODO(you): fill me in!
	algodToken := "0af285fbb066a145e38df913ad3ff637b76f1b178022a06215042ea620488a73" // TODO(you): fill me in!

	algodURL, err := url.Parse(localNodeURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot parse algod URL %s: %v\n", localNodeURL, err)
		os.Exit(1)
	}
	if len(*coinKey) == 0 {
		fmt.Fprintf(os.Stderr, "No master coin key passed, exiting.")
		os.Exit(1)
	}
	restClient := client.MakeRestClient(*algodURL, algodToken)

	results := make(map[string]uint64)
	curRound := *firstRound
	finalRound := *lastRound
	if curRound > finalRound {
		fmt.Fprintf(os.Stderr, "first round %d is after last round %d, exiting \n", curRound, finalRound)
		os.Exit(1)
	}

	sawInitializeMessage := false
	for {
		if curRound > finalRound {
			break
		}

		if *verboseFlag {
			fmt.Printf("Checking round %d...\n", curRound)
		}

		txns, err := restClient.TransactionsByAddr(*coinKey, curRound, curRound)
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
				var note examplecoin.NoteField
				err = dec.Decode(&note)
				if err != nil {
					break
				}

				switch note.Type {
				case examplecoin.NoteInitialize:
					if results, err = examplecoin.ProcessInitialize(results, note.Initialize, txn); err == nil {
						sawInitializeMessage = true
						if *verboseFlag {
							fmt.Printf("Saw an initialize message with supply %d.\n", note.Initialize.Supply)
						}
					} else {
						fmt.Printf("Error processing initialize message %v - err was \"%v\". Attempting to continue anyways.", note.Initialize, err)
					}
				case examplecoin.NoteTransfer:
					if results, err = examplecoin.ProcessTransfer(results, note.Transfer, txn); err == nil {
						if *verboseFlag {
							fmt.Printf("Saw a transfer message from %s to %s of amount %d", note.Transfer.Source, note.Transfer.Destination, note.Transfer.Amount)
						}
					} else {
						fmt.Printf("Error processing transfer message %v - err was \"%v\". Attempting to continue anyways.", note.Transfer, err)
					}
				default:
					continue
				}
			}
		}
		curRound++
	}

	if !sawInitializeMessage {
		fmt.Fprintf(os.Stderr, "Did not find an examplecoin initialization from key %s\n", *coinKey)
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
