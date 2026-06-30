# Project Rules

## Documentation Language Mirror

English documentation is the source of truth. When creating or modifying English documentation, also create or update the corresponding Chinese translation under the nearest `zh/` documentation directory.

Chinese translation directories mirror their English document locations:

- `docs/adr/*.md` -> `docs/zh/adr/*.md`
- `specs/*.md` -> `specs/zh/*.md`
- `domain/market/*.md` -> `domain/market/zh/*.md`

Note: do not read the project's Chinese documentation when gathering project context or deciding what to change. Use the English documentation as the source, then write the Chinese translation to match it.

## Root Documentation

The root `README.md` is user-facing Chinese documentation. Models and automation agents must not read the root `README.md` for project context.

Use the root `INDEX.md` as the model-facing project entry point.

When modifying the root `INDEX.md`, update the root `README.md` in the same change so user-facing documentation stays aligned.

## Code Style

Prefer extracting cohesive helper functions over writing large functions. Keep functions focused on one responsibility, and split complex control flow into named functions when doing so improves readability and testability.

Add code comments in Chinese. Comments should explain non-obvious intent, constraints, or domain rules; avoid comments that merely repeat what the code already says.
