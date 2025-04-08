#!/bin/sh
EMAIL=$1
PASSWORD=$2
if [ -z "$1" ]; then
  echo "第1引数が指定されていません。\nEmailにはtest@example.comを使用します。"
  EMAIL=test@example.com
fi

if [ -z "$2" ]; then
  echo "第2引数が指定されていません。\nPasswordにはpasswordを使用します。"
  PASSWORD=password
fi

RESULT=$(curl -s -X POST localhost:1323/signin \
  -H "Content-Type: application/json" \
  -d '{"email":"'$EMAIL'","password":"'$PASSWORD'"}' \
  | sed 's/.*"token":"\([^"]*\)".*/\1/')
echo $RESULT
# トークンが取得できたか確認するのだ
if [ -z "$RESULT" ]; then
  echo "トークンの取得に失敗しました。"
  exit 1
fi

# 環境変数に設定するのだ
export TOKEN="$RESULT"
echo "以下を実行してください\nexport TOKEN=$TOKEN"