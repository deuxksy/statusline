@./.ai/RULES.md

# Antigravity CLI Specific Guidelines
- Ensure all tool outputs and markdown links use absolute file URIs.
- Statusline integration test: `echo '{"antigravity":"v1"}' | ./statusline --cli=antigravity`.
- Quota integration test: `cat docs/samples/antigravity.json | ./statusline --cli=antigravity`.
