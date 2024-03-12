#!/usr/bin/env sh

mkdir -p ./builds/darwin
fyne package -os darwin -icon ./logo.png -name CSV文件礼包码拆分工具
mv CSV文件礼包码拆分工具.app ./builds/darwin

