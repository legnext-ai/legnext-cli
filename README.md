# Legnext CLI (`legnext.ai`)

**Midjourney & Multimodal API CLI for AI agents ([https://legnext.ai](https://legnext.ai)).** One command -> image/video. JSON by default. Built for Claude Code, Codex, Cursor, Cline, and shell scripts.

- **Official Platform & API Keys**: [https://legnext.ai](https://legnext.ai)
- **API Documentation & Endpoints**: [https://docs.legnext.ai](https://docs.legnext.ai)
- **Pricing & Plans**: [https://legnext.ai/pricing](https://legnext.ai/pricing)
- **Midjourney Prompt Library**: [https://legnext.ai/prompts](https://legnext.ai/prompts)

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

| Command | Description |
|---|---|
| `imagine "<prompt>"` | generate a 2048x2048 4-grid via [Legnext Midjourney API](https://legnext.ai) |
| `upscale <task_id> <n>` | upscale grid cell 0-3 |
| `variation <task_id> <n>` | variation of cell 0-3 (`--remix` to rewrite prompt) |
| `describe <url>` | reverse-engineer 4 prompts from a public image URL |
| `status <task_id>` | one-shot task state, no polling |
| `balance` | account credits and points |
| `login` / `init` | save API key · install Claude Code skill |

Blocks until done (auto-polls). `--text` for human output, `--no-wait` for async, `--out <path>` to save.

## License

MIT
