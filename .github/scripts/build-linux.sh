# Build linux
go env -w GOOS=linux GOARCH=amd64
go build -o ./build/mcsm-linux .