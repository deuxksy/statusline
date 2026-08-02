# statusline

`statusline`은 AI 코딩 어시스턴트(Antigravity, Claude Code, Codex)를 위한 고성능 단일 행 상태바 생성기입니다. Go 언어로 작성되었으며, STDIN을 통해 JSON 상태 스트림을 입력받아 5ms 이내의 빠른 속도로 실시간 실행 문맥을 렌더링합니다.

---

## 🚀 빠른 시작 (Quick Start)

```bash
# 바이너리 빌드
go build -o statusline ./cmd/statusline

# Antigravity 샘플 페이로드로 테스트 실행
cat docs/samples/antigravity.json | ./statusline --cli=antigravity
```

---

## 📚 문서 인덱스 (Diátaxis Framework)

`statusline`의 모든 문서는 [Open Knowledge Framework (OKF)](./docs/okf/README.md) 표준에 따라 Diátaxis 프레임워크 구조로 정돈되어 있습니다.

### 🎓 튜토리얼 (Tutorials)
- [튜토리얼 개요](./docs/okf/tutorials/index.md) - [시작하기 (Getting Started)](./docs/okf/tutorials/getting-started.md): 빠른 설치, 기본 사용법 및 파이프라인 구성.

### 🛠️ 사용 방법 안내 (How-To Guides)
- [사용 가이드 개요](./docs/okf/how-to-guides/index.md) - [레이아웃 및 요소 설정](./docs/okf/how-to-guides/configure-elements.md), [신규 CLI 어댑터 추가](./docs/okf/how-to-guides/add-new-cli-adapter.md) 가이드.

### 📖 레퍼런스 (Reference)
- [문서 허브 (Docs Hub)](./docs/README.md) - 전체 문서 디렉토리 구조 및 역할 안내.
- [OKF 스펙 허브](./docs/okf/README.md) - OKF 기술 문서 중앙 인덱스.
- [STDIN 페이로드 샘플](./docs/samples/README.md) - CLI 호스트별 입력 샘플 및 테스트 명령어.
- [레퍼런스 개요](./docs/okf/reference/index.md) - [CLI 페이로드 명세](./docs/okf/reference/cli-payload-spec.md), [설정 파일 스키마](./docs/okf/reference/configuration-schema.md), [아키텍처 레퍼런스](./docs/okf/reference/architecture.md).

### 💡 설명/개념 (Explanation)
- [개념 문서 개요](./docs/okf/explanation/index.md) - [설계 결정 사항](./docs/okf/explanation/design-decisions.md), [Diátaxis 프레임워크](./docs/okf/explanation/diataxis-framework.md).

---

## 📄 라이선스 (License)

MIT License.
