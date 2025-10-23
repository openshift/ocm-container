#!/bin/bash

i=0
while [[ $i -lt 10 ]]
do
  echo $i
  ((i++))
  sleep 2
done
