#!/bin/sh
if [ -z "$TOKEN" ]; then
  echo "Error: 環境変数TOKENが設定されていません。\nexport TOKEN=xxxで登録してから実行してください。"
  exit 1
fi


curl -X GET localhost:1323/api/todo \
  -H "Authorization: Bearer $TOKEN"