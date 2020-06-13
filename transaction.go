package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

type transaction interface {
	fmt.Stringer
	register(durations chan string)
	getReader() io.Reader
	getID() string
	isReady() bool
}

type regularTransaction struct {
	amount uint
	to     string
	user   user
	rpcID  uint
}

type sensorTransaction struct {
	regularTransaction
	txID     string
	isReadyM bool
}

func (rtx *regularTransaction) String() string {
	return fmt.Sprintf("{\"jsonRPC\": \"2.0\", \"id\": %d, \"method\": \"avm.send\", \"params\": {\"amount\": %d, \"assetID\": \"AVA\", \"to\": \"%s\", \"username\": \"%s\", \"password\": \"%s\"}}", rtx.rpcID, rtx.amount, rtx.to, rtx.user["name"], rtx.user["pass"])
}

func (rtx *regularTransaction) getReader() io.Reader {
	return strings.NewReader(rtx.String())
}

func (rtx *regularTransaction) getID() string { return "" }
func (stx *sensorTransaction) getID() string  { return stx.txID }

func (rtx *regularTransaction) isReady() bool { return true }
func (stx *sensorTransaction) isReady() bool  { return stx.isReadyM }

func (rtx *regularTransaction) register(durations chan string) { /*noop*/ }
func (stx *sensorTransaction) register(durations chan string) {
	start := time.Now()
	duration := stx.getID() + "," + strconv.FormatUint(uint64(stx.rpcID), 10)
	localdurations := make(chan string, len(queryAddressesList))

	for i := 0; i < len(queryAddressesList); i++ {
		go func(iCopy int) {
			for true {
				status, err := getTxStatus(stx, queryAddressesList[iCopy])
				if err != nil {
					fmt.Fprintf(os.Stderr, "Could not get status of transaction %s from %s: %v", stx.getID(), queryAddressesList[iCopy], err)
					localdurations <- "ERR"
					break
				}
				if status == "Accepted" || status == "Rejected" {
					end := time.Now()
					localdurations <- fmt.Sprintf("%s,%f", status, end.Sub(start).Seconds())
					break
				}
			}
		}(i)
	}

	go func() {
		first := false
		for i := 0; i < len(queryAddressesList); i++ {
			if first {
				duration = <-localdurations
				first = false
			} else {
				duration = fmt.Sprintf("%s,%s", duration, <-localdurations)
			}
		}
		stx.isReadyM = true
		durations <- duration
	}()
}
