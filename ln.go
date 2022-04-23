package main

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"log"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/macaroons"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"gopkg.in/macaroon.v2"
)

func createMacaroon(raw []byte) (*macaroon.Macaroon, error) {
	mac := &macaroon.Macaroon{}
	if err := mac.UnmarshalBinary(raw); err != nil {
		return nil, fmt.Errorf("unmarshal macaroon: %w", err)
	}

	return mac, nil
}

func createCredentials(raw []byte) (credentials.TransportCredentials, error) {
	if raw == nil {
		return credentials.NewTLS(nil), nil
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(raw) {
		return nil, errors.New("could not create TLS certs from bytes")
	}

	return credentials.NewClientTLSFromCert(certPool, ""), nil
}

func dialLnd(ctx context.Context, target string, mac, certBytes []byte) (*grpc.ClientConn, error) {
	log.Printf("dialing to LND at=%s", target)

	rpcCreds, err := createMacaroon(mac)
	if err != nil {
		return nil, err
	}

	transportCreds, err := createCredentials(certBytes)
	if err != nil {
		return nil, err
	}

	lndConn, err := grpc.DialContext(ctx, target,
		grpc.WithBlock(),
		grpc.WithTransportCredentials(transportCreds),
		grpc.WithReturnConnectionError(),
		grpc.FailOnNonTempDialError(true),
		grpc.WithPerRPCCredentials(
			macaroons.MacaroonCredential{Macaroon: rpcCreds},
		),
	)

	if err != nil {
		return nil, fmt.Errorf("could not dial to lnd: %w", err)
	}

	lnd := lnrpc.NewLightningClient(lndConn)

	res, err := lnd.GetInfo(ctx, &lnrpc.GetInfoRequest{})
	if err != nil {
		return nil, fmt.Errorf("lnd getinfo: %w", err)
	}

	log.Printf("connected to LND \tchain=%s \tversion=%s \theight=%d",
		res.Chains[0].Network, res.Version, res.BlockHeight)

	return lndConn, nil
}
