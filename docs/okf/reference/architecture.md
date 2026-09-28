# Reference: Statusline Cumulative Architecture

이 문서는 `statusline`의 점진적·누적 재활용(Cumulative Architecture) 구조를 정의합니다. CLI 단계에서 구축된 어댑터, 수집기, 데이터 모델은 일회성으로 버려지지 않고, GUI 에이전트 수집기와 OS 트레이 위젯, Wails 대시보드에 100% 동일 부품으로 재사용됩니다.

---

## 1. 전체 아키텍처 다이어그램 (Mermaid)

```mermaid
graph TD
    subgraph SOURCELAYER [1. 데이터 입력 소스 - Data Sources]
        subgraph CLISRC [CLI 에이전트]
            CLI_AGY[Antigravity CLI]
            CLI_CLAUDE[Claude Code]
        end
        subgraph GUISRC [GUI 에이전트]
            GUI_IDE[Antigravity IDE]
            GUI_VSC[VS Code 확장]
            GUI_CODEX[Codex Work]
            GUI_CLAUDE[Claude Work]
        end
        subgraph CLOUDSRC [클라우드 플랫폼]
            API_OAI[OpenAI Platform API]
            API_ANTH[Anthropic API]
            API_GCP[Google Cloud Billing API]
        end
    end

    subgraph CORELAYER [2. 공통 모니터링 엔진 - Core Engine]
        subgraph INGEST [수집 및 파싱 레이어]
            ADAPTER[Engine Adapters - Antigravity Claude Codex Z.AI]
            COLLECTOR[Collectors & Watchers - Local DB Log RPC]
        end

        MODEL[UnifiedStatus 도메인 모델]
        VCS[Git VCS Enrichment - 10ms budget]
    end

    subgraph CACHELAYER [3. 공유 상태 및 캐시 브릿지]
        CACHE_LIVE[live_session.json - 실시간 활성 세션 스냅샷]
        CACHE_ACC[accounts.json - 계정별 쿼터 플랜 누적 비용]
    end

    subgraph PRESENTATION [4. 프레젠테이션 레이어 - Pluggable UI]
        OUT_CLI[터미널 상태바 - Lipgloss ANSI 단일행]
        OUT_TRAY[OS 시스템 트레이 위젯 - systray]
        OUT_WAILS[리치 플로팅 대시보드 - Wails Webview]
    end

    CLI_AGY -->|STDIN Hook| ADAPTER
    CLI_CLAUDE -->|STDIN Hook| ADAPTER

    GUI_IDE -->|세션 로그 감시| COLLECTOR
    GUI_VSC -->|workspaceStorage 감시| COLLECTOR
    GUI_CODEX -->|app-server RPC 또는 로컬 DB| COLLECTOR
    GUI_CLAUDE -->|세션 이벤트 감시| COLLECTOR

    API_OAI -->|Admin API 폴링| COLLECTOR
    API_ANTH -->|Usage API 폴링| COLLECTOR
    API_GCP -->|Billing API 폴링| COLLECTOR

    ADAPTER --> MODEL
    COLLECTOR --> MODEL
    VCS --> MODEL

    MODEL -->|인메모리 즉시 전달| OUT_CLI
    MODEL -->|브릿지 덤프| CACHE_LIVE
    COLLECTOR -->|주기적 동기화| CACHE_ACC

    CACHE_LIVE -->|변경 감지| OUT_TRAY
    CACHE_ACC -->|계정 데이터 로드| OUT_TRAY

    CACHE_LIVE -->|IPC 또는 WebSocket| OUT_WAILS
    CACHE_ACC -->|실시간 렌더링| OUT_WAILS
```

---

## 2. 레이어별 역할 및 재활용 원칙

1. **데이터 소스 (Inputs)**:
   - CLI의 STDIN 훅, GUI의 로컬 로그/DB 감시, 클라우드 공식 API를 입력 채널로 둡니다.
2. **코어 엔진 (Core Engine)**:
   - 모든 입력은 `internal/adapter`와 `internal/collect`를 거쳐 단일 포맷인 `model.UnifiedStatus` 구조체로 표준화됩니다.
   - 이 코어 로직은 CLI, 백그라운드 데몬, 위젯 앱 어디서나 동일하게 import되어 재사용됩니다.
3. **공유 상태 브릿지 (Bridge)**:
   - 터미널 CLI는 초저지연(<5ms)으로 화면을 그리면서 동시에 `~/.cache/statusline/live_session.json`에 스냅샷을 덤프합니다.
   - 위젯과 대시보드는 이 표준 캐시 파일을 구독(Consume)하여 화면을 갱신합니다.
4. **프레젠테이션 (Outputs)**:
   - UI 껍데기만 Lipgloss(CLI) ➔ systray(트레이) ➔ Wails(웹뷰)로 교체되며, 비즈니스 로직과 데이터 파이프라인은 100% 공유됩니다.

