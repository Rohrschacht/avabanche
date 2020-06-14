package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	flag "github.com/spf13/pflag"
)

var queryAddressesList []string

const (
	modeFinal = iota
	modeCreat
	modeInitU
	modeBurst
)

var mode = modeFinal

func main() {
	// flags
	numtxs := flag.UintP("num", "n", 10, "number of generated transactions")
	sensorEvery := flag.UintP("sensor-every", "e", 5, "send a sensor transaction every N transactions")
	sensorRate := flag.Float64P("sensor-rate", "r", 0, "fraction between 0 and 1. determines the fraction of transactions that are sensors. in conflict with 'sensor-every'. if that is set explicitly, 'sensor-rate' will be ignored")

	amount := flag.UintP("amount", "a", 10, "amount of nAVA to be sent per transaction")
	address1 := flag.StringP("address1", "x", "", "one of the addresses involved in the transactions")
	address2 := flag.StringP("address2", "y", "", "one of the addresses involved in the transactions")

	username1 := flag.StringP("user1", "u", "", "username that owns first address")
	pass1 := flag.StringP("password1", "p", "", "password for user1")

	username2 := flag.String("user2", "", "username that owns second address. if not provided, oneway method is assumed")
	pass2 := flag.String("password2", "", "password for user2")

	outFileName := flag.StringP("outfile", "o", "", "output file for the sensor transactions")
	queryAddresses := flag.StringP("query-addresses", "q", "127.0.0.1:9650", "comma seperated list of addresses to query against. only the first one will be used for issuing transactions, the rest will be used for sensors")

	oneway := flag.BoolP("oneway", "1", false, "tokens will only be sent from address1 to address2 and not back")

	modeflag := flag.StringP("mode", "m", "finalization", "set the benchmark mode to: finalization, createusers, initusers or burst")

	numusers := flag.Uint("numusers", 0, "number of users to create in createusers mode")
	usersfile := flag.String("usersfile", "", "file containing the user information for createusers, initusers and burst mode")

	flag.Parse()

	sensorEveryExplicit := false

	flag.Visit(func(f *flag.Flag) {
		if f.Name == "sensor-every" {
			sensorEveryExplicit = true
		}
	})

	// fmt.Printf("%d\n%d\n%f\n%d\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%t\n%s\n%t\n%s\n", *numtxs, *sensorEvery, *sensorRate, *amount, *address1, *address2, *username1, *pass1, *username2, *pass2, *outFileName, sensorEveryExplicit, *queryAddresses, *oneway, *modeflag)

	// sanity checks
	modeoptions := map[string]bool{
		"finalization": true,
		"createusers":  true,
		"initusers":    true,
		"burst":        true,
	}
	if !modeoptions[*modeflag] {
		fmt.Fprintf(os.Stderr, "Mode should be one of: finalization, createusers or burst!\n")
		os.Exit(1)
	}
	if *modeflag == "finalization" {
		mode = modeFinal
	}
	if *modeflag == "createusers" {
		mode = modeCreat
	}
	if *modeflag == "initusers" {
		mode = modeInitU
	}
	if *modeflag == "burst" {
		mode = modeBurst
	}

	if mode == modeCreat {
		createUsers(*numusers, *usersfile)
	}
	if mode == modeInitU {
		initUsers(*usersfile)
	}

	if mode == modeBurst && *usersfile == "" {
		fmt.Fprintf(os.Stderr, "Cannot run burst mode without a usersfile!\n")
		os.Exit(1)
	}

	if mode == modeFinal && (*address1 == "" || *address2 == "") {
		fmt.Fprintf(os.Stderr, "Please specify the two addresses to use for the transactions!\n")
		os.Exit(1)
	}

	if mode == modeFinal && (*username1 == "" || *pass1 == "") {
		fmt.Fprintf(os.Stderr, "Please specify user and password for address1!\n")
		os.Exit(1)
	}

	if mode == modeFinal && (*username2 == "" && *pass2 == "") {
		*oneway = true
	} else if !*oneway {
		fmt.Fprintf(os.Stderr, "Please specify the username and password for address2!\n")
		os.Exit(1)
	}

	if *sensorRate < 0 || *sensorRate > 1 {
		fmt.Fprintf(os.Stderr, "Please specify a sensor rate between 0 and 1!\n")
		os.Exit(1)
	}

	if !strings.Contains(*queryAddresses, ":") {
		fmt.Fprintf(os.Stderr, "Please specify at least one node address with port like this: -q 127.0.0.1:9650\n")
		os.Exit(1)
	}

	// logic
	queryAddressesList = strings.Split(*queryAddresses, ",")
	user1 := user{"name": *username1, "pass": *pass1}
	user2 := user{"name": *username2, "pass": *pass2}
	if *sensorRate != 0 && !sensorEveryExplicit {
		*sensorEvery = uint(float64(*numtxs) * *sensorRate)
	}
	numsensors := uint(math.Ceil(float64(*numtxs) / float64(*sensorEvery)))
	durations := make(chan string, numsensors)
	users := loadUsers(*usersfile)

	outputfunc := func(str string) { fmt.Println(str) }
	if *outFileName != "" {
		outfile, err := os.OpenFile(*outFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not write to file: %s: %v\n", *outFileName, err)
			os.Exit(4)
		}
		defer outfile.Close()
		writer := bufio.NewWriter(outfile)
		outputfunc = func(str string) {
			writer.WriteString(fmt.Sprintf("%s\n", str))
			writer.Flush()
		}
	}

	if mode == modeFinal {
		balance, err := getBalance(*address1)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not get balance for %s: %v\n", *address1, err)
			os.Exit(2)
		}

		if *oneway && balance < uint64(*numtxs**amount) {
			fmt.Fprintf(os.Stderr, "It seems your account has insufficient funds: %d < %d * %d\n", balance, *numtxs, *amount)
			os.Exit(2)
		} else if !*oneway && balance < uint64(*numtxs**amount/2) {
			fmt.Fprintf(os.Stderr, "It seems your account has insufficient funds: %d < %d * %d/2\n", balance, *numtxs, *amount)
			os.Exit(2)
		}

		*sensorEvery = 1
		numsensors = *numtxs
	}

	outputfunc(fmt.Sprintf("BEGIN,%s,%d", time.Now().Format(time.RFC3339Nano), *numtxs))

	go func() {
		var factory transactionFactory
		if mode == modeFinal {
			factory = transactionFactory{amount: *amount, to: *address2, user: user1}
		}
		for i := uint(0); i < *numtxs; i++ {
			if mode == modeBurst {
				factory = transactionFactory{amount: *amount, to: users[i+1]["address"], user: users[i]}
			}

			if mode == modeFinal && !*oneway && i > (*numtxs/2) {
				factory.to = *address1
				factory.user = user2
			}

			var tx transaction
			if i%*sensorEvery == 0 {
				tx = factory.newSensorTransaction(i)
			} else {
				tx = factory.newRegularTransaction(i)
			}

			err := sendTx(tx)
			if err != nil {
				wentThrough := false
				for k := 0; k < 5; k++ {
					fmt.Fprintf(os.Stderr, "Could not send transaction: %v. Retrying...\n", err)
					err = sendTx(tx)
					if err == nil {
						wentThrough = true
						break
					}
				}
				if !wentThrough {
					fmt.Fprintf(os.Stderr, "Sending a transaction failed 5 times. Aborting!\n")
					os.Exit(3)
				}
			}

			tx.register(durations)

			if mode == modeFinal {
				for !tx.isReady() {
					time.Sleep(100 * time.Millisecond)
				}
			}
		}
	}()

	for i := uint(0); i < numsensors; i++ {
		duration := <-durations
		outputfunc(duration)
	}

	outputfunc(fmt.Sprintf("END,%s", time.Now().Format(time.RFC3339Nano)))
}
