# 설치

Lightning Multitool을 설치하는 방법에는 두 가지가 있습니다.

1.  **사전 빌드된 바이너리 사용(권장)**: GitHub에서 최신 릴리스를 다운로드합니다.
2.  **소스에서 빌드**: 리포지토리를 복제하고 직접 빌드합니다.

## 사전 빌드된 바이너리 사용

이것이 가장 쉬운 시작 방법입니다.

1.  **최신 릴리스 다운로드**

    [GitHub 릴리스](https://github.com/asheswook/lightning-multitool/releases/) 페이지로 이동하여 운영 체제에 맞는 바이너리를 다운로드합니다.

2.  **바이너리 실행**

    다운로드 후 터미널에서 실행할 수 있습니다.

    ```bash
    chmod +x lightning-multitool
    ```

    ```bash
    ./lightning-multitool
    ```

## 소스에서 빌드

소스에서 프로젝트를 빌드하려면 시스템에 Go가 설치되어 있어야 합니다.

1.  **리포지토리 복제**

    ```bash
    git clone https://github.com/asheswook/lnurl.git
    cd lnurl
    ```

2.  **프로젝트 빌드**

    ```bash
    go build ./cmd/server/main.go -o lightning-multitool
    ```

3.  **바이너리 실행**

    ```bash
    ./lightning-multitool
    ```
