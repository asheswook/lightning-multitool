# Configuration

The Lightning Multitool is configured using environment variables. For ease of use, you can create a `.env` file in the root of the project to store your configuration settings. The application will automatically load this file on startup.

To get started, you can copy the example configuration file:

```bash
cp .env.example .env
```

Then, edit the `.env` file with your desired settings.

---

## Application Config

These settings control the core behavior of the application server.

| Variable      | Description                                                                                             | Default       |
|---------------|---------------------------------------------------------------------------------------------------------|---------------|
| `SERVER_HOST` | The network interface the server listens on. Use `0.0.0.0` to listen on all available interfaces.         | `127.0.0.1`   |
| `SERVER_PORT` | The port the server listens on.                                                                         | `8080`        |
| `NODE_KIND`   | The kind of Lightning node to connect to. Currently, only `lnd` is supported.                           | `lnd`         |
| `DOMAIN`      | **Required.** The domain name for your Lightning Address and Nostr NIP-05 ID (e.g., `yourdomain.com`).      | (none)        |
| `USERNAME`    | **Required.** Your username for the Lightning Address (e.g., `satoshi`).                                    | (none)        |

---

## LND Connection Config

These settings are required to connect the tool to your LND node.

| Variable            | Description                                                                                                                              | Default                                                 |
|---------------------|------------------------------------------------------------------------------------------------------------------------------------------|---------------------------------------------------------|
| `LND_HOST`          | The host and port of your LND node's gRPC interface (e.g., `localhost:10009`).                                                           | `localhost:8080`                                        |
| `LND_MACAROON_PATH` | The full path to your LND `admin.macaroon` file. An invoice macaroon can be used, but it will limit functionality like Zaps.               | `~/.lnd/data/chain/bitcoin/mainnet/admin.macaroon`      |
| `LND_CERT_PATH`     | The full path to your LND `tls.cert` file. This is often not needed if the certificate is trusted by the system.                           | (none)                                                  |

---

## LNURL Config

Customize the behavior of LNURL payments.

| Variable            | Description                                                                                             | Default       |
|---------------------|---------------------------------------------------------------------------------------------------------|---------------|
| `MIN_SENDABLE_MSAT` | The minimum amount, in millisatoshis, that can be sent to your Lightning Address (1 sat = 1000 msat).    | `1000`        |
| `MAX_SENDABLE_MSAT` | The maximum amount, in millisatoshis, that can be sent in a single payment.                               | `1000000000`  |
| `COMMENT_ALLOWED`   | The maximum character length for comments in payment requests. Set to `0` to disable comments.          | `255`         |

---

## Nostr Config

Configure your Nostr identity for NIP-05 and Zaps.

| Variable            | Description                                                                                             | Default       |
|---------------------|---------------------------------------------------------------------------------------------------------|---------------|
| `NOSTR_PRIVATE_KEY` | **Required for Nostr features.** Your Nostr private key in hexadecimal format (64 characters). This is used to sign NIP-05 verification events. | (none)        |
