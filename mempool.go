package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type mempoolEntry struct {
	TXID string `json:"txid"`
}

func fetchMempoolEntries() ([]mempoolEntry, error) {
	const path = "https://mempool.space/api/mempool/recent"

	res, err := http.Get(path)

	defer res.Body.Close()
	if err != nil {
		return nil, err
	}

	var entries []mempoolEntry
	return entries, json.NewDecoder(res.Body).Decode(&entries)

}

func fetchTransaction(txid string) ([]mempoolSpaceVout, error) {
	path := fmt.Sprintf("https://mempool.space/api/tx/%s", txid)

	res, err := http.Get(path)
	defer res.Body.Close()
	if err != nil {
		return nil, err
	}

	var mempoolSpaceResult mempoolSpaceTx
	return mempoolSpaceResult.Vout, json.NewDecoder(res.Body).Decode(&mempoolSpaceResult)
}

type mempoolSpaceTx struct {
	Vin []struct {
		Prevout mempoolSpaceVout
	} `json:"vin"`
	Vout []mempoolSpaceVout `json:"vout"`
}

type mempoolSpaceVout struct {
	ScriptPubKey string `json:"scriptpubkey"`
	Address      string `json:"scriptpubkey_address"`
}
