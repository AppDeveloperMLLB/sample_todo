#!/bin/sh
if [ -z "$TOKEN" ]; then
  echo "Error: 環境変数TOKENが設定されていません。\nexport TOKEN=xxxで登録してから実行してください。"
  exit 1
fi

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

curl -X PUT localhost:1323/api/todo/$ID \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"title":"'$TITLE'","body":"'$BODY'"}'