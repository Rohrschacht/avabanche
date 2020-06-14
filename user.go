package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"
)

type user map[string]string

func generateUsers(numusers uint, usersfile string) {
	users := loadUsers(usersfile)

	file, err := os.OpenFile(usersfile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not open usersfile: %v!\n", err)
		os.Exit(4)
	}
	defer file.Close()

	writer := csv.NewWriter(file)

	for i := uint(len(users)); i < numusers; i++ {
		name := fmt.Sprintf("avabancheBurstuser_%d_%s", i, randomString(5))
		u := user{"name": name, "pass": randomString(10)}
		success, err := createUser(u)
		if err != nil || !success {
			fmt.Fprintf(os.Stderr, "There was an error creating user %v: %v\n", u, err)
			os.Exit(5)
		}

		writer.Write([]string{u["name"], u["pass"]})
		writer.Flush()
		if writer.Error() != nil {
			fmt.Fprintf(os.Stderr, "There was an error writing to usersfile: %v\n", writer.Error())
			os.Exit(6)
		}
	}
}

func initUsers(usersfile string) {
	users := loadUsers(usersfile)

	file, err := os.OpenFile(usersfile, os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not open usersfile: %v!\n", err)
		os.Exit(4)
	}
	defer file.Close()

	writer := csv.NewWriter(file)

	for _, u := range users {
		if u["address"] == "" {
			address, err := createAddress(u)
			if err != nil || address == "" {
				fmt.Fprintf(os.Stderr, "There was an error creating an address for user %v: %v\n", u, err)
				os.Exit(5)
			}

			writer.Write([]string{u["name"], u["pass"], address})
			writer.Flush()
			if writer.Error() != nil {
				fmt.Fprintf(os.Stderr, "There was an error writing to usersfile: %v\n", writer.Error())
				os.Exit(6)
			}
		} else {
			writer.Write([]string{u["name"], u["pass"], u["address"]})
			writer.Flush()
			if writer.Error() != nil {
				fmt.Fprintf(os.Stderr, "There was an error writing to usersfile: %v\n", writer.Error())
				os.Exit(6)
			}
		}
	}
}

func faucetUsers(usersfile string, amount uint, u user) {
	users := loadUsers(usersfile)

	for _, ur := range users {
		if ur["address"] == "" {
			fmt.Fprintf(os.Stderr, "There are uninitialized users in the usersfile!\n")
			os.Exit(4)
		}
	}

	balance, err := getBalance(u["address"])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not get balance for %s: %v\n", u["address"], err)
		os.Exit(2)
	}

	if uint64(uint(len(users))*amount) > balance {
		fmt.Fprintf(os.Stderr, "Address does not have enough funds: %d * %d > %d!\n", len(users), amount, balance)
		os.Exit(2)
	}

	factory := transactionFactory{amount: amount, user: u}
	durations := make(chan string, len(users))

	for i, ur := range users {
		factory.to = ur["address"]
		tx := factory.newSensorTransaction(uint(i))

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
		for !tx.isReady() {
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func loadUsers(usersfile string) []user {
	file, err := os.OpenFile(usersfile, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not open usersfile: %v!\n", err)
		os.Exit(4)
	}
	defer file.Close()

	users := []user{}

	reader := csv.NewReader(file)
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			if !strings.Contains(err.Error(), "wrong number of fields") {
				fmt.Fprintf(os.Stderr, "There was an error reading the usersfile: %v\n", err)
				os.Exit(4)
			}
		}

		if len(record) == 2 {
			users = append(users, user{"name": record[0], "pass": record[1]})
		} else if len(record) == 3 {
			users = append(users, user{"name": record[0], "pass": record[1], "address": record[2]})
		} else {
			fmt.Fprintf(os.Stderr, "There is an error in the formatting of the usersfile!\n")
			os.Exit(4)
		}
	}

	return users
}

func randomString(length uint) string {
	rand.Seed(time.Now().UnixNano())
	const characterPool = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	buf := make([]byte, length)
	for i := range buf {
		buf[i] = characterPool[rand.Intn(len(characterPool))]
	}
	return string(buf)
}
