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
- [시작하기 (Getting Started)](./docs/okf/tutorials/getting-started.md) - 빠른 설치, 기본 사용법 및 파이프라인 구성.

### 🛠️ 사용 방법 안내 (How-To Guides)
- [레이아웃 및 요소 설정](./docs/okf/how-to-guides/configure-elements.md) - 위젯 커스텀, 색상 설정 및 tmux 통합.
- [신규 CLI 어댑터 추가](./docs/okf/how-to-guides/add-new-cli-adapter.md) - 새로운 AI 에이전트 호스트 지원 확장.

### 📖 레퍼런스 (Reference)
- [문서 허브 (Docs Hub)](./docs/README.md) - 전체 문서 디렉토리 구조 및 역할 안내.
- [OKF 스펙 허브](./docs/okf/README.md) - OKF 기술 문서 중앙 인덱스.
- [STDIN 페이로드 샘플](./docs/samples/README.md) - CLI 호스트별 입력 샘플 및 테스트 명령어.
- [CLI 페이로드 명세](./docs/okf/reference/cli-payload-spec.md) - STDIN JSON 페이로드 상세 스키마.
- [설정 파일 스키마](./docs/okf/reference/configuration-schema.md) - `config.json` 구성 및 임계값 옵션.
- [아키텍처 레퍼런스](./docs/okf/reference/architecture.md) - 파이프라인 구조, 인터페이스 및 실행 흐름.

### 💡 설명/개념 (Explanation)
- [설계 결정 사항](./docs/okf/explanation/design-decisions.md) - 주요 기술적 선택, 성능 한계 및 안심 실패(fail-soft) 설계.
- [Diátaxis 프레임워크](./docs/okf/explanation/diataxis-framework.md) - 문서 구조화 프레임워크 도입 배경.

---

## 📄 라이선스 (License)

MIT License.
