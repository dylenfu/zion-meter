#!/bin/bash

split(){
  echo ""
  echo ""
  echo ""
  sleep 1s
}

# group number
group=$1
user=$2
last=$3
echo "group number $group and accounts per group $user and last time $last"

# kill program and backup log
echo "close existing meter..."
./stop.sh
ssh -p 32000 ubuntu@10.203.0.43 "cd /data/zion/tps && ./stop.sh"
split

# show ulimit -n which denote max open file number
echo "show max tcp connections..."
maxConn=$(ulimit -n)
echo "allow max tcp connections $maxConn"
split

# distribute execute binary file
echo "distibute binary file..."
md5sum meter
scp -P 32000 meter ubuntu@10.203.0.43:/data/zion/tps/meter
ssh -p 32000 ubuntu@10.203.0.43 "cd /data/zion/tps && md5sum meter"
split

# distribute config file
echo "distibute config file..."
md5sum config.json
scp -P 32000 config.json ubuntu@10.203.0.43:/data/zion/tps/config.json
ssh -p 32000 ubuntu@10.203.0.43 "cd /data/zion/tps && md5sum config.json"
split

# run and show log
echo "tps testing start..."
nohup make run group=$group user=$user last=$last >> tps.log 2>&1 &
sleep 60s;
echo "mache1 started!"
ssh -p 32000 ubuntu@10.203.0.43 "cd /data/zion/tps && nohup ./start.sh $group $user $last"
sleep 60s;
echo "mache2 started!"
tail -f tps.log