### statik -src=assets/dist -f
export GOARCH=amd64
export GOOS=linux
go build -o Builds/v1.0/ACH-1.0.0alpha14 ach
export GOOS=windows
go build -o Builds/v1.0/ACH-1.0.0alpha14.exe ach

