@./.ai/RULES.md

# Claude Code Specific Guidelines
- Use TDD for all Go feature additions and adapter implementations.
- Statusline command integration test: `echo '{"omcLabel":"OMC"}' | ./statusline --cli=claude`.
- Sample payload test: `cat docs/samples/claude.json | ./statusline --cli=claude`.
