# lightning-multitool

`lightning-multitool` is an easy-to-use and lightweight multitool for people who operate personal Lightning nodes.

## Warning: This project is currently working but documentation and code quality need improvement (Proof of Concept). I keep improving this project and will be tidied up soon.

## Features

- Create a Lightning Address for your domain (e.g., `you@yourdomain.com`).
- Receive Lightning payments (Zaps) via your Nostr profile.
- Link your Nostr public key to your domain with NIP-05 support.
- (Planned) Remotely control your wallet using Nostr Wallet Connect (NIP-47).

## Getting Started

[See this documentation](https://lightning-multitool.pages.dev).

## Implementation

### LNURL

- [x] [LUD-01: Base LNURL encoding and decoding](https://github.com/lightningnetwork/luds/blob/master/lud-01.md)
- [x] [LUD-06: BIP32-based seed generation for auth protocol](https://github.com/lightningnetwork/luds/blob/master/lud-06.md)
- [x] [LUD-12: Comments in payRequest](https://github.com/lightningnetwork/luds/blob/master/lud-12.md)
- [x] [LUD-16: Paying to static internet identifiers](https://github.com/lightningnetwork/luds/blob/master/lud-16.md)

### Nostr

- [x] [NIP-05: Mapping Nostr keys to DNS-based internet identifiers](https://github.com/nostr-protocol/nips/blob/master/05.md)
- [ ] [NIP-47: Nostr Wallet Connect](https://github.com/nostr-protocol/nips/blob/master/47.md)
- [x] [NIP-57: Lightning Zaps](https://github.com/nostr-protocol/nips/blob/master/57.md)

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
