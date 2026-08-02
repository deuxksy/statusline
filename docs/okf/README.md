# statusline v0.2 개발자 및 사용자 문서 (OKF)

`statusline` v0.2 프로젝트의 **Open Knowledge Framework (OKF)** 문서 허브입니다.

본 문서는 개발자와 사용자의 탐색 목적에 맞춰 [Diátaxis 프레임워크](https://diataxis.fr/) 체계에 따라 4개의 카테고리로 정리되어 있습니다.

---

## 🧭 Diátaxis 문서 탐색 체계

```
                  학습 중심 (Learning)
                           │
             튜토리얼     │    사용 방법 안내
           (Tutorials)    │    (How-To Guides)
                          │
  실무 중심 ──────────────┼────────────── 이론 중심
  (Practical Work)        │               (Theoretical Knowledge)
                          │
            레퍼런스      │      설명 및 개념
           (Reference)    │      (Explanation)
                          │
                 정보 중심 (Information)
```

### 1. 🎓 [튜토리얼 (Tutorials)](./tutorials/index.md) *(학습 중심)*
`statusline`을 처음 접하는 사용자를 위한 단계별 입문 가이드입니다.
- [getting-started.md](./tutorials/getting-started.md): 설치, 초기 설정 및 첫 번째 상태바 파이프라인 실행 방법.

### 2. 🛠️ [사용 방법 안내 (How-To Guides)](./how-to-guides/index.md) *(실무 중심)*
구체적인 작업 목표를 달성하기 위한 실용적인 안내서입니다.
- [configure-elements.md](./how-to-guides/configure-elements.md): 레이아웃 요소 커스텀, 색상 설정 및 tmux 연동 방법.
- [add-new-cli-adapter.md](./how-to-guides/add-new-cli-adapter.md): 새로운 AI 에이전트 CLI 호스트 어댑터 구현하기.

### 3. 📖 [레퍼런스 (Reference)](./reference/index.md) *(정보 중심)*
기술 명세, 스키마 및 아키텍처 상세 정보입니다.
- [cli-payload-spec.md](./reference/cli-payload-spec.md): Antigravity (`agy`), Claude Code, Codex용 STDIN JSON 페이로드 상세 스키마.
- [configuration-schema.md](./reference/configuration-schema.md): `config.json` 구성 항목, 레이아웃 속성 및 임계값 옵션.
- [architecture.md](./reference/architecture.md): 패키지 구조, 인터페이스 정의 및 데이터 처리 파이프라인.

### 4. 💡 [설명 및 개념 (Explanation)](./explanation/index.md) *(이해 중심)*
설계 배경, 기술적 선택 이유 및 성능 구조 안내입니다.
- [design-decisions.md](./explanation/design-decisions.md): 단일 행 렌더링, 안심 실패(fail-soft) 설계 및 5ms 이내 응답 보장 설계 결정.
- [diataxis-framework.md](./explanation/diataxis-framework.md): OKF v0.2에서 Diátaxis 프레임워크를 도입한 이유와 활용법.
