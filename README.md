# user-manager

## Running Locally

```bash
make build-server
./bin/server run --config configs/server/config.yaml
```

You can then connect to the DB.

```bash
sqlite3 /tmp/user_manager.db
```

You can then hit the endpoint.

```bash
curl http://localhost:8080/v1/users/api_keys

curl http://localhost:8080/v1/users/api_keys \
  --request POST \
  --data '{"name": "test-key"}'
```
