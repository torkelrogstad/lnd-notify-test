# lnd-notify-test

Reproduction repo for issue [#6443](https://github.com/lightningnetwork/lnd/issues/6443) in LND. 

1. Connect to LND
2. Fetch a TX in the mempool
3. Subscribe to updates about it using the `chainnotifier.RegisterConfirmationsNtfn` API

Expected outcome: seeing updates about the TX being printed once it receives confirmations. 

# Usage

```shell
$ go build .
$ ./lnd-notify-test -cert=BASE64_CERT -macaroon=BASE64_MACAROON -lnd=LND_HOST_WITH_PORT
```