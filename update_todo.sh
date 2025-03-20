#!/bin/sh
if [ -z "$1" ]; then
  echo "Error: 第1引数が指定されていません。"
  exit 1
fi

if [ -z "$2" ]; then
  echo "Error: 第2引数が指定されていません。"
  exit 1
fi

if [ -z "$3" ]; then
  echo "Error: 第2引数が指定されていません。"
  exit 1
fi

ID=$1
TITLE=$2
BODY=$3

curl -X PUT localhost:1323/todo/$ID -H "Content-Type: application/json" -d '{"title":"'$TITLE'","body":"'$BODY'"}'