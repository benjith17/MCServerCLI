# Build macOS (intel)
GOOS=darwin GOARCH=amd64 go build -o ./build/mcsm-osx-intel .


# Build macOS (arm64)
GOOS=darwin GOARCH=arm64 go build -o ./build/mcsm-osx-arm .