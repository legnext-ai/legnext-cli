# Legnext CLI

**Midjourney CLI for AI agents.** One command → image. JSON by default. Built for Claude Code, Cursor, Cline, and shell scripts.

```bash
go install github.com/legnext-ai/legnext-cli@latest
legnext login
legnext imagine "a photorealistic cat on a windowsill" --out ./cat.png
```

## Commands

| | |
|---|---|
| `imagine "<prompt>"` | generate a 2048×2048 4-grid |
| `upscale <task_id> <n>` | upscale grid cell 0-3 |
| `variation <task_id> <n>` | variation of cell 0-3 (`--remix` to rewrite prompt) |
| `describe <url>` | reverse-engineer 4 prompts from a public image URL |
| `status <task_id>` | one-shot task state, no polling |
| `balance` | account credits and points |
| `login` / `init` | save API key · install Claude Code skill |

Blocks until done (auto-polls). Pass `--text` for human output, `--no-wait` for async, `--out <path>` to save.

Homebrew, npm, and `curl | sh` packages coming soon.

## License

MIT
