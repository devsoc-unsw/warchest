# Warchest

## Development
### Starting development
1. Start all services and prepare a dev container (app, money-module, Postgres, TigerBeetle):
```bash
docker compose up -d
```
If this doesn't work install the docker-compose plugin

2. To enter the dev environment
```bash
docker exec -it warchest-dev-container bash
```

3. Go linting
```bash
golangci-lint run
```

### Closing down development
1. To exit the dev environment
```bash
exit
```

2. To close all services
```bash
docker compose down
```
use `bun run lint` for lint and `bun run test` in the dev shell for tests.
tests need to end in the suffix .test.ts.


