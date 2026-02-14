# JSON Translation Editor

A self-hosted tool for managing JSON translation files.

## Features.

1.  **Visual Editor**: Interactive UI for editing translation keys.
2.  **Missing Key Tracking**: Filter to show only missing translations.
3.  **Validation**: Ensures placeholders (e.g., `{user}`) are preserved.
4.  **Auto Translate**: AI-powered translation for missing fields.
5.  **Example Mode**: Try the editor without creating a project.

## AI Translation

The "Auto Translate" feature uses the OpenAI API to automatically fill missing translation fields.

-   **Model**: `gpt-4o`
-   **Configuration**: Requires `OPENAI_API_KEY` environment variable.
-   **Cost Efficiency**: Only missing fields are sent to the API.

### Setup
Add your API key to the `.env` file:
```bash
OPENAI_API_KEY=sk-...
```

## Logs & Metrics & Monitoring

Metrics are on new port with opentel compatiple style.
Logs and metrics are gathered with Alloy and sent to loki & prometheus.

## Future Work / Todo

- [ ] Fix UI glitch: validation error persists after correction until reload.

## Versioning

This project uses automatic semantic versioning powered by anothrNick/github-tag-action
.

By default:

Every merge/push to main bumps the patch version (0.0.x)

Unless a version flag is explicitly defined in the commit message

🔢 How Version Bumping Works

The action determines the next version based on commit message flags.

If no flag is found, it falls back to the configured default (usually patch or minor depending on your setup).

📍 Manual (Commit Message–Driven)

You can control the version bump by including one of the following keywords in your commit message:

Flag	Result	Example
#major	1.2.3 → 2.0.0	Breaking API changes
#minor	1.2.3 → 1.3.0	New backwards-compatible feature
#patch	1.2.3 → 1.2.4	Bug fix
#none	No new tag created	Docs / CI changes
🧾 Example Commit Messages
git commit -m "feat: add OAuth login #minor"


→ bumps 1.4.2 → 1.5.0

git commit -m "fix: handle nil pointer panic #patch"


→ bumps 1.4.2 → 1.4.3

git commit -m "refactor!: rewrite auth system #major"


→ bumps 1.4.2 → 2.0.0

git commit -m "docs: update README #none"


→ no new version tag
