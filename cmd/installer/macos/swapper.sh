#!/bin/bash
#Start RenEx Atomic Swapper only if it is not running
if [ "$(ps -ef | grep -v grep | grep swapper | wc -l)" -le 0 ] 
then
 # Note starting swapper not as a sudoer
 ~/.swapper/bin
 echo "RenEx Atomic Swapper Started"
else
 echo "RenEx Atomic Swapper Already Running"
fi