# Project Rules

## Documentation Language Mirror

English documentation is the source of truth. When creating or modifying English documentation, also create or update the corresponding Chinese translation under the nearest `zh/` documentation directory.

Chinese translation directories mirror their English document locations:

- `docs/adr/*.md` -> `docs/zh/adr/*.md`
- `specs/*.md` -> `specs/zh/*.md`
- `domain/market/*.md` -> `domain/market/zh/*.md`

Note: do not read the project's Chinese documentation when gathering project context or deciding what to change. Use the English documentation as the source, then write the Chinese translation to match it.
