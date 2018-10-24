#!/bin/bash
. ~/.bashrc

today=`date -d "0 day" +%Y%m%d`
workDir=$(cd $(dirname $0); pwd)

logFile="${LOG_PATH}/spider_${today}.log"

## 开始执行
cd ${workDir}
rm -fr ./main
go build ./src/main/

./main > ${logFile}

rm -fr ./main