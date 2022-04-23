package main

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"flag"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/chainrpc"
)

var (
	lnd         = flag.String("lnd", "", "LND endpoint")
	rawMacaroon = flag.String("macaroon", "", "base64-encoded macaroon")
	rawCert     = flag.String("cert", "", "base64-encoded TLS cert")
)

func realMain() error {
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt)
		sig := <-signals

		log.Printf("received signal: %s", sig)
		// cancel the context, this should cause things to
		// exit gracefully
		cancel()
	}()

	macBytes, err := base64.StdEncoding.DecodeString(*rawMacaroon)
	if err != nil {
		return err
	}

	var certBytes []byte
	if *rawCert != "" {
		certBytes, err = base64.StdEncoding.DecodeString(*rawCert)
		if err != nil {
			return err
		}
	}

	lnCtx, lnCancel := context.WithTimeout(ctx, time.Second*5)
	defer lnCancel()

	lnClient, err := dialLnd(lnCtx, *lnd, macBytes, certBytes)
	if err != nil {
		return err
	}

	lnd := lnrpc.NewLightningClient(lnClient)
	info, err := lnd.GetInfo(ctx, &lnrpc.GetInfoRequest{})
	if err != nil {
		return err
	}

	log.Printf("best block height: %d", info.BlockHeight)

	mempool, err := fetchMempoolEntries()
	if err != nil {
		return err
	}

	tx, err := fetchTransaction(mempool[0].TXID)
	if err != nil {
		return err
	}

	txid, err := hex.DecodeString(mempool[0].TXID)
	if err != nil {
		return err
	}

	scriptPubKey, err := hex.DecodeString(tx[0].ScriptPubKey)
	if err != nil {
		return err
	}

	notifier := chainrpc.NewChainNotifierClient(lnClient)

	confRequest := &chainrpc.ConfRequest{
		Txid:       txid,
		NumConfs:   1,
		HeightHint: info.BlockHeight - 6,
		Script:     scriptPubKey,
	}

	log.Printf("subscribing to TXID=%s and scriptPubKey=%s",
		hex.EncodeToString(txid), hex.EncodeToString(scriptPubKey))

	sub, err := notifier.RegisterConfirmationsNtfn(ctx, confRequest)
	if err != nil {
		return err
	}

	defer sub.CloseSend()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-sub.Context().Done():
			return sub.Context().Err()
		default:
		}

		update, err := sub.Recv()
		if err != nil {
			return err
		}

		log.Printf("GOT UPDATE: %+v", update)
	}
}

func main() {
	if err := realMain(); err != nil {
		log.Fatal(err)
	}

	log.Println("exiting")
}
