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

3. Dingo linting
```bash
dingo lint ./...
```

4. Dingo transpiling
```bash
dingo go ./...
```

5. Running tests (note files should be transpiled already)
```bash
go test -v ./...
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


