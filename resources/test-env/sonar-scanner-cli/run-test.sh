#!/bin/bash

# 閥值
threshold=$1

# 檢查參數是否提供
if [ -z "$threshold" ]; then
  echo "Usage: $0 <threshold>"
  exit 1
fi

sonar_sourceEncoding=${SONAR_SOURCE_ENCODING:-"UTF-8"}
sonar_sources=$SONAR_SRC
sonar_token=$SONAR_TOKEN
sonar_host=$SONAR_HOST_URL
sonar_project_key=$(echo $(echo $RANDOM | md5sum | cut -c 1-8)$(echo $(date +%s) | md5sum | cut -c 1-8))

# echo "projectKey: $sonar_project_key"
# echo "sources: $sonar_sources"
# echo "sourceEncoding: $sonar_sourceEncoding"
# echo "host URL: $sonar_host"
# echo "token: $sonar_token"

# write sonar-project.properties
echo "sonar.projectKey=$sonar_project_key" > sonar-project.properties
echo "sonar.sources=$sonar_sources" >> sonar-project.properties
echo "sonar.sourceEncoding=$sonar_sourceEncoding" >> sonar-project.properties
echo "sonar.host.url=$sonar_host" >> sonar-project.properties
echo "sonar.token=$sonar_token" >> sonar-project.properties

echo "[Judge] Starting SonarQube analysis..."

/usr/bin/entrypoint.sh && while ! curl -s -H "Authorization: Bearer $sonar_token" "$sonar_host/api/ce/task?id=$(cat /tmp/.scannerwork/report-task.txt | grep "ceTaskId" | cut -d'=' -f2)" | grep -q '"status":"SUCCESS"'; do sleep 1; done && curl -s -H "Authorization: Bearer $sonar_token" "$sonar_host/api/issues/search?componentKeys=$sonar_project_key&types=CODE_SMELL" | grep -o '"total":[0-9]*' | cut -d':' -f2 | head -n1 | while read count; do echo "[Judge] Found $count code smells (limit: $threshold)" && if [ "$count" -gt $threshold ]; then echo "fail" && exit 1; else echo "pass"; fi; done

curl -X POST -H "Authorization: Bearer $sonar_token" "$sonar_host/api/projects/delete?project=$sonar_project_key"
