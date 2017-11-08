#!/bin/sh

workdir=$(cd $(dirname $0)/; pwd)

pids=`ps xo pid,cmd | awk '{print $1,$2}'  | grep "task" | awk '{print $1}'`

if [ ! -n "$pids" ]
then
	echo "Process does not exist"
	exit 1
fi

for pid in $pids
do
	path=`ls -l /proc/$pid | sed -n 11p | awk '{print $NF}'`
	if [ "$path" == "$workdir" ]
	then
		kill $pid
	fi
done

echo "Stop process success"
