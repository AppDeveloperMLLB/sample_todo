#!/bin/sh
if [ -z "$1" ]; then
  echo "Error: 第1引数が指定されていません。"
  exit 1
fi

if [ -z "$2" ]; then
  echo "Error: 第2引数が指定されていません。"
  exit 1
fi

TITLE=$1
BODY=$2

curl -X POST localhost:1323/todo -H "Content-Type: application/json" -d '{"title":"'$TITLE'","body":"'$BODY'"}'