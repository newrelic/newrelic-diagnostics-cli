#!/bin/bash
#script can optionally take flags
#-p port
#-l location (resource path)
#-e expected response


#initialize variables with defaults
port='8080'
location='/'
expected_response='200'

#get command line args
while getopts p:l:e: flag
do
    case "${flag}" in
        p) port=${OPTARG};;
        l) location=${OPTARG};;
        e) expected_response=${OPTARG};;
    esac
done

WAIT_COUNT=0
until [ "$WAIT_COUNT" -gt 30 ] || [ "`curl --silent --show-error --connect-timeout 1 -I http://localhost:$port$location | grep -m 1 $expected_response`" ]; 
do 
    sleep 1;
    WAIT_COUNT=$(($WAIT_COUNT+1))
done
