# 라이트닝 주소

라이트닝 주소는 라이트닝 네트워크를 통해 결제를 받는 데 사용할 수 있는 간단하고 사람이 읽을 수 있는 주소입니다. 이메일 주소(예: `you@yourdomain.com`)처럼 보입니다.

## 작동 방식

Lightning Multitool은 이 주소를 표준 LNURL로 변환하는 프로세스를 처리하며, 호환되는 모든 지갑에서 이를 사용하여 결제할 수 있습니다.

## 설정

라이트닝 주소를 설정하려면 `.env` 파일에서 다음 환경 변수를 구성해야 합니다.

- `DOMAIN`: 소유한 도메인 이름(예: `mydomain.com`).
- `USERNAME`: 원하는 사용자 이름(예: `pororo`).

설정이 완료되면 라이트닝 주소는 `USERNAME@DOMAIN` (예: `pororo@mydomain.com`)이 됩니다.
