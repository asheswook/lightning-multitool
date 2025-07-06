# Nostr 지원

Lightning Multitool은 Nostr과 통합하여 Zaps 및 NIP-05 신원 확인을 포함한 몇 가지 주요 기능을 제공합니다.

## Zaps (NIP-57)

Zaps는 Nostr 사용자에게 라이트닝 결제를 보내는 방법입니다. Lightning Multitool을 사용하면 자신의 라이트닝 노드로 직접 Zaps를 받을 수 있습니다.

## NIP-05 확인

NIP-05를 사용하면 Nostr 공개 키를 도메인에 연결하여 검증 가능한 신원을 제공할 수 있습니다. 이를 통해 다른 사람들이 귀하의 Nostr 프로필을 더 쉽게 찾고 신뢰할 수 있습니다.

## 설정

Nostr 기능을 활성화하려면 다음 환경 변수를 설정해야 합니다.

- `NOSTR_PRIVATE_KEY`: 16진수 형식의 Nostr 개인 키입니다.

이를 통해 도구는 NIP-05 확인과 같은 기능을 위해 귀하를 대신하여 Nostr 이벤트에 서명할 수 있습니다.
