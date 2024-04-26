#!/bin/bash

curl 'http://localhost:17000/?cmd=reset'

curl 'http://localhost:17000/?cmd=white'

curl 'http://localhost:17000/?cmd=figure+0.0+0.0'

len=100
step=$(bc -l <<< "scale=2; 1/$len")
while true; do
	for ((i=0; i <= len; i++)); do
		curl "http://localhost:17000/?cmd=move+0$step+0$step"
		curl 'http://localhost:17000/?cmd=update'
	done
	for ((i=0; i <= len; i++)); do
		curl "http://localhost:17000/?cmd=move+-0$step+-0$step"
		curl 'http://localhost:17000/?cmd=update'
	done
done
