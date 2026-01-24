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

Logs are gathered with Loki and Alloy.
Metrics with prometheus.

## Future Work / Todo

- [ ] Fix UI glitch: validation error persists after correction until reload.