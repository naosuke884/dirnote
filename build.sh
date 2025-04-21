VERSION=$1
if [ -z "$VERSION" ]; then
  echo "Error: VERSION is not provided."
  echo "Usage: $0 <VERSION>"
  exit 1
fi
GOOS=linux CGO_ENABLE=0 go build -trimpath -ldflags='-s -w' -o bin/${VERSION}dirnote-linux main.go
GOOS=windows CGO_ENABLE=0 go build -trimpath -ldflags='-s -w' -o bin/${VERSION}/dirnote-windows.exe main.go
GOOS=darwin CGO_ENABLE=0 go build -trimpath -ldflags='-s -w' -o bin/${VERSION}/dirnote-darwin main.go