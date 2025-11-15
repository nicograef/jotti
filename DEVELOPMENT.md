```sh
docker compose down --volumes
docker compose up postgres migrate backend --build
```

```bash
psql -h localhost -p 5432 -U ${POSTGRES_USER} -d jotti
```
