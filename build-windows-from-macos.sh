#!/usr/bin/env sh

mkdir -p ./builds/windows
env GOOS="windows" GOARCH="amd64" CGO_ENABLED="1" CC="x86_64-w64-mingw32-gcc" fyne package -os windows -icon ./logo.png -name CSV文件礼包码拆分工具
mv CSV文件礼包码拆分工具.exe ./builds/windows

#env GOOS="windows" GOARCH="amd64" CGO_ENABLED=false fyne package -os windows -icon ./logo.png -name CSV文件礼包码拆分工具
