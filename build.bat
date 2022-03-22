statik -src=assets/build/ -f
set GOOS=linux
set GOARCH=amd64
go build -o Releases/v1.0/ACH-1.0.0alpha1 ach
set GOOS=windows
go build -o Releases/v1.0/ACH-1.0.0alpha1.exe ach
pause