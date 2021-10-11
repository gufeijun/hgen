#/bin/bash
if [ ! -d "./bin" ]; then
    mkdir bin
fi

oldgoos=$(go env GOOS)

go env -w GOOS=linux
go build -o hgen main.go
tar cvf bin/hgen-linux64.tar.gz hgen

go env -w GOOS=darwin
go build -o hgen main.go
tar cvf bin/hgen-mac64.tar.gz hgen
rm hgen

go env -w GOOS=windows
go build -o hgen.exe main.go
zip bin/hgen-win64.zip hgen.exe
rm hgen.exe

go env -w GOOS=$oldgoos
