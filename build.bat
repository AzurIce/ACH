statik -src=assets/dist -f
set GOARCH=amd64
set GOOS=linux
go build -o Builds/v1.0/ACH-1.0.0alpha12 ach
set GOOS=windows
go build -o Builds/v1.0/ACH-1.0.0alpha12.exe ach