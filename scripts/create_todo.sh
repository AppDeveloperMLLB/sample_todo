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

TITLE=$1
BODY=$2

curl -X POST localhost:1323/api/todo \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"title":"'$TITLE'","body":"'$BODY'"}'