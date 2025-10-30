## schema migrations with pgroll

```bash
go install github.com/xataio/pgroll@latest

pgroll init --postgres-url postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:5432/jotti?sslmode=disable

psql -h localhost -p 5432 -U ${POSTGRES_USER} -d jotti

psql -h localhost -p 5432 -U ${POSTGRES_USER} -d jotti -f ./database/01-initial-schema.sql
```

```bash
curl -i -X POST http://localhost:3000/create-user \
  -H "Content-Type: application/json" \
  -d '{ "username": "nico", "password": "svj" }'
```


```bash
curl -i -X POST http://localhost:3000/login \
  -H "Content-Type: application/json" \
  -d '{ "username": "nico", "password": "svj" }'
```
