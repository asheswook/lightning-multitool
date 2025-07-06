# Nostr Support

Lightning Multitool integrates with Nostr to provide several key features, including Zaps and NIP-05 identity verification.

## Zaps (NIP-57)

Zaps are a way to send Lightning payments to Nostr users. With Lightning Multitool, you can receive Zaps directly to your own Lightning node.

## NIP-05 Verification

NIP-05 allows you to link your Nostr public key to your domain, providing a verifiable identity. This makes it easier for others to find and trust your Nostr profile.

## Configuration

To enable Nostr features, you need to set the following environment variable:

- `NOSTR_PRIVATE_KEY`: Your Nostr private key in hex format.

This allows the tool to sign Nostr events on your behalf for features like NIP-05 verification.
