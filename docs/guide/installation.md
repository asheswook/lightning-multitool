# Installation

There are two ways to install the Lightning Multitool:

1.  **Using a pre-built binary (Recommended)**: Download the latest release from GitHub.
2.  **Building from source**: Clone the repository and build it yourself.

## Using a Pre-built Binary

This is the easiest way to get started.

1.  **Download the latest release**

    Go to the [GitHub Releases](https://github.com/asheswook/lightning-multitool/releases/) page and download the binary for your operating system.

2.  **Run the binary**

    After downloading, you can run the binary from your terminal:

    ```bash
    ./lmt
    ```

## Building from Source

If you prefer to build the project from source, you'll need to have Go installed on your system.

1.  **Clone the repository**

    ```bash
    git clone https://github.com/asheswook/lightning-multitool
    cd lightning-multitool
    ```

2.  **Build the project**

    ```bash
    go build -o lmt ./cmd/server/main.go
    ```

3.  **Run the binary**

    ```bash
    ./lmt
    ```
