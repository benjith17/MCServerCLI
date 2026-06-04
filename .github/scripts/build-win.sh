# Build windows
go-winres make --arch amd64

go env -w GOOS=windows GOARCH=amd64
go build -o ./build/mcsm-win64.exe .