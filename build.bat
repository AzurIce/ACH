statik -src=assets/build/ -f
set GOOS=linux
set GOARCH=amd64
go build -o Builds/v1.0/ACH-1.0.0alpha3 ach
set GOOS=windows
go build -o Builds/v1.0/ACH-1.0.0alpha3.exe ach
pause