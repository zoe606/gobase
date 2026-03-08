Generate code from a migration and wire everything up.

Ask the user for the MIGRATION number if not provided as an argument, then run:

1. `make gen-full MIGRATION=$MIGRATION` - Generate all layers (entity, dto, repo, usecase, handler)
2. `make wire` - Auto-wire DI, routes, and contracts
3. `make check-all` - Verify everything compiles and passes
4. `make swag` - Regenerate Swagger docs

Report results after each step. Stop if any step fails.
