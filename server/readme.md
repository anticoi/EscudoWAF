Documentación swagger: http://localhost:2333/swagger/index.html

Regenerar la documentación swagger:

`go install github.com/swaggo/swag/cmd/swag@latest`
`swag init --dir ./,../pkg`
`swag fmt --dir ./,../pkg`

Formatear código:
`gofmt -w .`
