package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func sendRequest(address, url string, data io.Reader) (*http.Response, error) {
	return http.Post(fmt.Sprintf("http://%s%s", address, url), "application/json", data)
}

func sendTx(tx transaction) error {
	resp, err := sendRequest(queryAddressesList[0], "/ext/bc/X", tx.getReader())
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("node returned status code: %d", resp.StatusCode)
	}

	if stx, ok := tx.(*sensorTransaction); ok {
		result, err := unmarshalResult(resp.Body)
		if err != nil {
			return err
		}

		var txID string
		err = json.Unmarshal(result["txID"], &txID)
		if err != nil {
			return err
		}

		stx.txID = txID
		fmt.Println("Transaction sent: ", stx.txID)
	}

	return nil
}

func getBalance(address string) (uint64, error) {
	response, err := sendRequest(queryAddressesList[0], "/ext/bc/X", strings.NewReader(fmt.Sprintf("{\"jsonrpc\": \"2.0\", \"id\": 1, \"method\": \"avm.getBalance\", \"params\": {\"address\": \"%s\", \"assetID\": \"AVA\"}}\n", address)))
	if err != nil {
		return 0, err
	}

	if response.StatusCode != 200 {
		return 0, fmt.Errorf("node returned status code: %d", response.StatusCode)
	}

	result, err := unmarshalResult(response.Body)
	if err != nil {
		return 0, err
	}

	var balanceStr string
	err = json.Unmarshal(result["balance"], &balanceStr)
	if err != nil {
		return 0, err
	}

	return strconv.ParseUint(balanceStr, 10, 64)
}

func getTxStatus(tx transaction, queryAddress string) (string, error) {
	if tx.getID() == "" {
		return "", fmt.Errorf("tx has no txID")
	}

	response, err := sendRequest(queryAddress, "/ext/bc/X", strings.NewReader(fmt.Sprintf("{\"jsonrpc\": \"2.0\", \"id\": 1, \"method\": \"avm.getTxStatus\", \"params\": {\"txID\": \"%s\"}}\n", tx.getID())))
	if err != nil {
		return "", err
	}

	if response.StatusCode != 200 {
		return "", fmt.Errorf("node returned status code: %d", response.StatusCode)
	}

	result, err := unmarshalResult(response.Body)
	if err != nil {
		return "", err
	}

	var status string
	err = json.Unmarshal(result["status"], &status)
	if err != nil {
		return "", err
	}

	return status, nil
}

func unmarshalResult(data io.ReadCloser) (map[string]json.RawMessage, error) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(data)
	defer data.Close()
	if err != nil {
		return nil, err
	}

	var responseJSON map[string]json.RawMessage
	err = json.Unmarshal(buf.Bytes(), &responseJSON)
	if err != nil {
		return nil, err
	}

	var result map[string]json.RawMessage
	err = json.Unmarshal(responseJSON["result"], &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
