# Legnext CLI

**Midjourney CLI for AI agents.** Generate Midjourney images from the command line — built for Claude Code, Cursor, Cline, and shell scripts.

## Status

Early development. Commands and flags may change before `v1.0`.

## Install

Coming soon. Planned distributions:

- Homebrew: `brew install legnext-ai/tap/legnext`
- npm: `npm install -g @legnext/cli`
- Shell: `curl -fsSL https://legnext.ai/install.sh | sh`

## Quick start

```bash
# one-time auth (paste API key from https://legnext.ai/app/api-keys)
legnext login

# generate an image (blocks until done, prints JSON result)
legnext imagine "a photorealistic cat on a windowsill"
```

## Design

- **JSON by default** — pass `--text` for human-readable output
- **Blocks by default** — CLI auto-polls upstream; pass `--no-wait` to return a task ID immediately
- **Local files as first-class inputs** — `legnext describe @./photo.png` uploads automatically
- **Env var auth** — `LEGNEXT_API_KEY=...` for CI and agent sandboxes

## License

MIT
