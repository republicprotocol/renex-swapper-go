#!/bin/sh
#Start RenEx Atomic Swapper only if it is not running
started=$(ps -ef | grep -v grep | grep swapper | wc -l)
if [ $(($(ps -ef | grep -v grep | grep swapper | wc -l))) = 0 ];
then
    ~/.swapper/bin
    echo "RenEx Atomic Swapper Started"
elif [ $((started)) = 1 ]
then
    echo "RenEx Atomic Swapper Already Running"
else
    echo "something wrong"    
fi