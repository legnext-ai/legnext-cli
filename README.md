# Legnext CLI

**Midjourney CLI for AI agents.** One command → image. JSON by default. Built for Claude Code, Cursor, Cline, and shell scripts.

```bash
# Homebrew
brew install legnext-ai/tap/legnext

# or npm
npx @legnext/cli imagine "hello world"

# or curl
curl -fsSL https://raw.githubusercontent.com/legnext-ai/legnext-cli/main/install.sh | sh

# then
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

Blocks until done (auto-polls). `--text` for human output, `--no-wait` for async, `--out <path>` to save.

## License

MIT
