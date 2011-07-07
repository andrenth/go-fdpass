#!/bin/sh

d=`dirname $0`
echo $d
cd $d/mycat && gomake && cd -
cd $d/openfile && gomake && cd -

cd $d
./mycat/mycat ./mycat/mycat.go
