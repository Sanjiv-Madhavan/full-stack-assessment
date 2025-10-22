package api

//go:generate go tool oapi-codegen --config ./config-server.yaml --package api -o ./server.gen.go ../../api/openapi.yml
//go:generate go tool oapi-codegen --config ./config-models.yaml --package scheme -o ../scheme/scheme.gen.go ../../api/openapi.yml
