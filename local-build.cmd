@echo off

@REM set PREFIX=%USERPROFILE%\Downloads
set PREFIX=D:\Softwares

set GOROOT=%PREFIX%\go1.21.0.windows-amd64\go
set GOPATH=%PREFIX%\gopath

set PATH=^
%PREFIX%\PortableGit\bin;^
%PREFIX%\winlibs-x86_64-posix-seh-gcc-11.2.0-mingw-w64-9.0.0-r1\mingw64\bin;^
%GOROOT%\bin;

set CGO_ENABLED=1
go build main.go &&^
pause
