---
name: legnext
description: Generate, upscale, vary, and describe Midjourney images via the Legnext CLI. Use whenever the user asks to create, generate, make, remix, upscale, or describe images, illustrations, photos, mockups, avatars, hero art, or any visual asset. All output is JSON by default for easy piping.
user_invocable: true
---

# Legnext CLI — Midjourney image generation for agents

Generate Midjourney images from the command line. The CLI blocks until
the task finishes (auto-polls upstream), outputs JSON by default, and
never exposes upstream provider details.

## Prerequisites

- `legnext` binary on PATH (`brew install legnext-ai/tap/legnext`, `curl -fsSL https://legnext.ai/install.sh | sh`, or `go install github.com/legnext-ai/legnext-cli@latest`)
- API key configured (user runs `legnext login` once, or exports `LEGNEXT_API_KEY`)
- Smoke-test auth: `legnext balance`

## Commands

All commands print JSON to stdout by default. Pass `--text` for a compact
human-readable form. Pass `--no-wait` to return immediately with a `job_id`.

### Generate — `imagine`

```bash
legnext imagine "<prompt>" [--ar 16:9] [--out ./image.png]
```

- Blocks ~30-90s (fast mode). No manual polling needed.
- Returns `{job_id, status, output.image_url (2048×2048 4-grid), output.image_urls[4], cost_points, duration_seconds, path}`
- `--ar` appends `--ar <value>` to the Midjourney prompt
- `--mode relax|turbo` appends `--relax` or `--turbo` to the prompt
- If `--out` is a file path, the grid is saved there; if a directory, the URL basename is used

### Upscale one cell — `upscale`

```bash
legnext upscale <task_id> <image_no>   # image_no: 0-3
```

Grid layout: `0 1 / 2 3` (top-left, top-right, bottom-left, bottom-right).

Flags: `--type subtle|creative` (default subtle), `--out`, `--no-wait`.

### Variation — `variation`

```bash
legnext variation <task_id> <image_no> [--type subtle|strong] [--remix "new prompt"]
```

### Describe — `describe`

```bash
legnext describe "https://example.com/image.png" [--text]
```

Returns 4 Midjourney-style prompts inferred from the image. URL must be
publicly reachable HTTP(S); local file paths are not yet supported.

### Status & balance

```bash
legnext status <task_id>   # one-shot task read, no polling
legnext balance            # account credits / points
```

## Typical workflows

**Generate then upscale the best cell:**

```bash
task_id=$(legnext imagine "a photorealistic red apple on a wooden table" --ar 1:1 | jq -r .job_id)
legnext upscale "$task_id" 2 --out ./apple.png
```

**Generate several options, pick one, remix it:**

```bash
task=$(legnext imagine "futuristic city skyline at dusk, synthwave, --ar 16:9" | jq -r .job_id)
legnext variation "$task" 0 --remix "add flying cars" --out ./city.png
```

**Extract a prompt from a reference image:**

```bash
legnext describe "https://example.com/style.png" --text
```

## Cost (approximate, in points)

| Command | Cost |
|---|---|
| `imagine` (fast) | 80 |
| `upscale` | 120 |
| `variation` | 80 |
| `describe` | 20-40 |

Check balance before long batches: `legnext balance | jq .available_points`.

## Error handling

- `no API key`: user needs `legnext login` or `export LEGNEXT_API_KEY=...`
- `HTTP 400 ... banned`: Midjourney rejected the prompt; rephrase and retry
- `HTTP 429`: rate limited; wait 30s and retry
- `timeout waiting for task`: long prompt / relax mode — pass `--timeout 20m`

## Output conventions

- Top-level `cost_points` and `duration_seconds` are always present on
  completed task output
- Upstream provider identity, internal worker IDs, and webhook config
  are stripped from all responses — you will never see them
- `output.image_url` is the 4-grid; `output.image_urls[]` has the 4
  individual cell URLs when the task type produces a grid
