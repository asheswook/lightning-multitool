# Lightning Address

A Lightning Address is a simple, human-readable address that you can use to receive payments over the Lightning Network. It looks like an email address (e.g., `you@yourdomain.com`).

## How it Works

The Lightning Multitool handles the process of converting this address into a standard LNURL, which can be used by any compatible wallet to pay you.

## Configuration

To set up your Lightning Address, you need to configure the following environment variables in your `.env` file:

- `DOMAIN`: Your domain name (e.g., `yourdomain.com`).
- `USERNAME`: Your desired username (e.g., `you`).

Once configured, your Lightning Address will be `USERNAME@DOMAIN`.
