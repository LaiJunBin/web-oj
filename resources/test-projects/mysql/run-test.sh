#!/bin/bash

# 傳入參數
test_no=$1

# 檢查參數是否提供
if [ -z "$test_no" ]; then
  echo "Usage: $0 <test_no>"
  exit 1
fi

# 動態設定檔案名稱
init_file="init-${test_no}.sql"
user_file="/app/user-${test_no}.sql"
answer_file="answer-${test_no}.sql"
container_name=$PROJECT_CONTAINER_NAME

# 檢查容器名稱是否設定
if [ -z "$container_name" ]; then
  echo "CONTAINER_NAME is not set"
  exit 1
fi

# 檢查檔案是否存在
if [ ! -f "$init_file" ] || [ ! -f "$user_file" ] || [ ! -f "$answer_file" ]; then
  echo "One or more required files (init-${test_no}.sql, user-${test_no}.sql, answer-${test_no}.sql) are missing."
  exit 1
fi

# 初始化資料庫
echo "Initializing database with $init_file..."
init_cmd='mysql --host='"$container_name"' -uroot --password=root --database=testdb -e "source '"$init_file"'"'
eval $init_cmd
if [ $? -ne 0 ]; then
  echo "Failed to initialize database."
  exit 1
fi

# 執行使用者的 SQL 並將結果存入變數
echo "Executing user SQL: $user_file..."
user_result=$(mysql --host="$container_name" -uroot --password=root --database=testdb -B -e "source $user_file;")
if [ $? -ne 0 ]; then
  echo "User SQL execution failed."
  exit 1
fi

# 執行標準答案 SQL 並將結果存入變數
echo "Executing judge SQL: $answer_file..."
answer_result=$(mysql --host="$container_name" -uroot --password=root --database=testdb -B -e "source $answer_file;")
if [ $? -ne 0 ]; then
  echo "Judge SQL execution failed."
  exit 1
fi

# 比較結果
echo "Comparing results..."
if [ "$user_result" == "$answer_result" ]; then
  echo "Test passed: User result matches the judge result."
  exit 0
else
  echo "Test failed: User result does not match the judge result."
  echo "User Result:"
  echo "$user_result"
  echo "Judge Result:"
  echo "$answer_result"
  exit 1
fi