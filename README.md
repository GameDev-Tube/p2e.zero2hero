# P2E - Predict to Earn

This is a prototype / PoC for Predict to Earn contract, frontend, and backend.

The main idea is - There are some trackers on a sports event, for example, if it's a boxing match, let's say
how many punches with left, with the right fist player a, and player b did, how many insults were thrown or how
many dramas on Twitter happened in a given window of time.

Having the prediction in the form of NFT allows for example trading them as the game progresses. If you lose
faith in your prediction, you could still recover by selling it, and if you believe that a given prediction
is underappreciated, you could buy it, giving the game more "fluidity", making it feel more lively.

# Setup and usage

## Dependencies for building

Included Makefile depends on:
- Golang (recommended 1.19 and higher, older versions might still work)
- Solc (^0.8.19). If you have the command as `solc`, alias it to `solc8` before building, or edit Makefile
- Abigen from Go-Ethereum.

All of these need to be in your `$PATH`.
It should be built on Linux.

## Building

Call `make`. If you want a delve build (i.e., for remote debugging), edit Makefile and uncomment `go build` with
additional flags, and comment out the "standard" `go build`.

## Configuring

First, we need to deploy the smart contract. There are a few ways to do it: you could take the ABI and BIN from
`solc-build` for deployment or paste the contract into Remix (along with dependencies) and use it for deployment.

It requires 2 params: BEP20 token that's the main token for purchases and Pancake Swap router 01.
On BSC testnet, Pancake address is `0xD99D1c33F9fC3444f8101754aBC46c52416550D1`.

Next, copy `config.json.dist` to `config.json`:
```bash
cp config.json.dist config.json
```

Edit config.json:
It's a json format file, in which all fields must be filled in. Fields:

`rpc_addr` (string) - URL to RPC for Binance Smart Chain or Binance Smart Chain testnet. Can be `https://` or `wss://`. Wss is recommended.
`psql_dsn` (string) - DSN to PostgreSQL. Example DSN to build from: `host=localhost user=dbuser password=dbpass dbname=dbname port=5432 sslmode=disable`
`indexer_start_block` (uint64) - Block since when the indexer starts scanning for logs. Insert here a block in which smart contract deployment was confirmed
`auth_key` (string) - Key to game referee panel. Referee must hold Owner of contract key in MetaMask. Since this is held as plaintext in config, I suggest a long string, [for example from here](https://grc.com/passwords.htm)
`p2e_contract` (string) - An Ethereum address (`0x[a-fA-F0-9]{40}`) for P2E smart contract
`bep20_contract` (string) - An Ethereum address for the underlying BEP20 token. In code, you may find reference to it as FAME token, as that's the target underlying token, but it should work on any other BEP20 token.
`listen` (string) - listen string for local HTTP server. It's `address:port`. To bind on all interfaces, the address should be `0.0.0.0`, to bind on localhost only, it should be `127.0.0.1`. Default is `127.0.0.1:8089`.
`external_url` (string) - Actual external URL string, i.e., for usage with Apache2 reverse proxy, we need to know this to build URLs. It MUST NOT contain `http[s]://` prefix, trailing slash, etc. Usually it will be for local testing same as `listen`, or FQDN (ServerName) of vhost.
`is_https` (bool) - does NOT control if the local HTTP server is HTTP or HTTPS, rather it controls building URLs. If you wrap the API with an SSL reverse proxy, set it up to `true`; otherwise, set it to `false`.

Transfer the config and server binary to the server that will host it, and set up Apache2 with proxy and SSL modules.
Get your hands on an SSL cert, which can be a Let's Encrypt free SSL. To help you with Apache2 configuration,
there is a vhost file included, with reverse proxy config.

You can add Pancake liquidity pool (V2!) for your selected tokens to other tokens, and use Remix to allow using
the given token for creating predictions. It's whitelist-based and allows users to create predictions using other
BEP20 tokens than what was provided in the constructor. Whitelisting will be picked up by the indexer and presented
on the frontend as an option. Hardcoded allowed slippage is 3%.

## Usage

Start up the server. Every 30 seconds, the indexer will report if it's behind. If you see failed ticks, it's normal,
as these usually are errors on RPC comms; it will retry internally. The reason for concern is if the logs are
spammed with this error, and no tick progresses.

During compilation, the HTML / CSS / JS files are embedded into the binary, so there is no need for a complicated setup
with FE/BE. Just reverse proxy it to wrap it with SSL, and it will have both FE and API.

Sometimes you will need to call one of 2 functions from Remix. Use the "At Address" feature of Remix for this.
These functions are:

1. Init Round - to start a new prediction round. It defines how much prediction costs, how much of it goes into
    the reward pool, consolidation reward pool, and fee pool.
2. Start Game - blocks minting new NFTs and emits an event after which the backend will start calculating leaderboards.
    In this phase, the referee should update sliders on `/game-panel`.

When the backend is ready (the logs about the indexer catching up stop appearing), initiate the round via Remix.
Now, let users create predictions or create them yourself for testing purposes.

When it comes to the event that predictions are about, for example, the boxing match is about to start,
run Start Game from Remix, and let the referee log into `/game-panel`. It starts a WebSocket connection and allows
real-time adjustments of results, changing the leaderboards. When the game is finalized, and the results are known
to be final (up to the referee), in the referee panel, press the "phase up button" button. A form shows up, in which
the referee inputs how many predictions should be selected for first, second, and third place. It's up to
the game rules if you want only one winning prediction, getting all the reward for that bracket, and many second
or third places, or any other combination.

After the round gets finalized, all NFTs generated from it can be burned for rewards by end users.
All burns give rewards, including "losing" predictions (consolidation prize).

# License

MIT.

# Production Ready?

**No.** This project is currently a prototype. Further development and testing are required before deploying it in a live environment.