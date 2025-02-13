#!/bin/bash

export GOOS=linux
export GOARCH=amd64
go build -o jito_proxy
chmod +x ./jito_proxy
git add .
git commit -m "update"
git push origin master
