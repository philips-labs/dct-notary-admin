# DCT Notary Admin Web

## Dependencies

Use Node.js 24 and the pnpm version declared in `package.json`.
The Docker build uses `node:24-alpine`, which includes Corepack; keep its Node
major aligned with `engines.node` and CI.
Install dependencies with `pnpm install --frozen-lockfile`, matching CI.
Keep `@fortawesome/fontawesome-svg-core` and `@fortawesome/free-solid-svg-icons`
pinned to the same patch version and update the lockfile together; different
versions can introduce incompatible `IconName` types.
Keep `vitest` and `@vitest/coverage-v8` pinned to the same exact version and
update them together. Mixed versions use incompatible coverage payload formats.

## Available Scripts

In the project directory, run:

### `pnpm start`

Runs the app in development mode at [http://localhost:5173](http://localhost:5173).

### `pnpm test`

Runs the Vitest test suite with coverage.

### `pnpm build`

Type-checks the app and builds the production bundle to `dist`.

### `pnpm preview`

Serves the production bundle locally after `pnpm build`.
