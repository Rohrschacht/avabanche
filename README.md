# avabanche

## About

This program is designed to interact with the [gecko
client](https://github.com/ava-labs/gecko) of the AVA platform. It is able to
perform the following actions:

- request the creation of users
- request X-Chain addresses for these users
- distribute AVA from a central address to the users
- submit transactions for X-Chain addresses
- track the time until a transaction is accepted by the network

Different modes are implemented that perform these actions. The goal is to
benchmark the performance of the AVA platform.

### <a name="Transaction_types"></a> Transaction types

A "sensorTransaction" is a construct used in this program that represents a
transaction whose time to acceptance is tracked. In order to do this, a
goroutine is created once the transaction is submitted that polls the gecko
client constantly for the status of the transaction.

In addition to that a "regularTransaction" will only get submitted to the gecko
client and not be tracked.

## <a name="Commandline_arguments"></a> Commandline arguments

```
Usage of avabanche:
-x, --address1 string one of the addresses involved in the transactions
-y, --address2 string one of the addresses involved in the transactions
-a, --amount uint amount of nAVA to be sent per transaction (default 1)
-m, --mode string set the benchmark mode to: finalization, createusers, initusers, faucetusers or burst (default "finalization")
-n, --num uint number of generated transactions (default 10)
--numusers uint number of users to create in createusers mode
-1, --oneway tokens will only be sent from address1 to address2 and not back
-o, --outfile string output file for the sensor transactions
-p, --password1 string password for user1
--password2 string password for user2
-q, --query-addresses string comma seperated list of addresses to query against. only the first one will be used for issuing transactions, the rest will be used for sensors (default "127.0.0.1:9650")
-e, --sensor-every uint send a sensor transaction every N transactions (default 5)
-r, --sensor-rate float fraction between 0 and 1. determines the fraction of transactions that are sensors. in conflict with 'sensor-every'. if that is set explicitly, 'sensor-rate' will be ignored
-u, --user1 string username that owns first address
--user2 string username that owns second address. if not provided, oneway method is assumed
--usersfile string file containing the user information for createusers, initusers, faucetusers and burst mode
```

## Modes

### finalization

In this mode, only [sensorTransactions](#Transaction_types) are generated. The
program will wait for the acceptance of each sensorTransaction before generating
the next one. This mode is intended to track the finalization time for
transactions over a long period of time.

The parameters supported for this mode are `-x, -y, -a, -n, -1, -o, -u, -p, --user2, --password2`. `-e` and `-r` have no effect, because every transaction
will be a sensorTransaction.

### createusers

This mode creates a specified amount of users with random user names and
passwords.

Supported parameters are `--numusers` and `--usersfile`. If there are already
users in the usersfile, only an amount of users needed to increase those up to
numusers is created. In other words, after running this mode, userfile will
contain at least numusers amount of users.

### initusers

This mode creates X-Chain addresses for every user in `--usersfile` who does not
have an address yet. `--usersfile` is the only valid parameter.

### faucetusers

This mode distributes some AVA across the users in the usersfile from a
specified X-Chain address.

Supported parameters are `--usersfile, -x, -u, -p` and `-a`, where `-a` does not
represent an amount of AVA to be sent per transaction, [as in the rest of the
modes](#Commandline_arguments), but rather the total amount of nAVA that should
be taken from `-x` and distributed among the users.

### burst

This mode generates many transactions at once, up to a maximum of the number of
users that are in the specified usersfile. Some of these transactions will be
sensorTransactions, according to the specified parameters `-e` or `-r`.

In order to run this mode, the `--usersfile` should have at least `-n` users in
it that were created using createusers, initusers and faucetusers and that each
have at least `-a` nAVA in their balance.

The transactions generated in this mode go from user 1 to user 2, from 2 to 3,
and from user `-n` back to user 1, completing a cycle. This way the burst mode
can be run multiple times consecutively.

Supported parameters are `--usersfile, -n, -a, -e, -r, -o`.

## Acknowledgements

This implementation of avabanche makes use of the following awesome open source
projects:

- [Go programming language](https://golang.org/)
- [spf13/pflag](https://github.com/spf13/pflag)
