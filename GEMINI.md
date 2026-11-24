@./.ai/RULES.md

# Antigravity CLI Specific Guidelines
- 대화에서 로컬 파일을 참조할 때는 절대 경로를 사용하고, 저장소 Markdown 링크에는 상대 경로를 사용한다.
- Statusline integration test: `echo '{"antigravity":"v1"}' | ./statusline --cli=antigravity`.
- Quota integration test: `cat docs/samples/antigravity.json | ./statusline --cli=antigravity`.
