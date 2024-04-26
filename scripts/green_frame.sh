curl 'http://localhost:17000/?cmd=white'

curl 'http://localhost:17000/?cmd=brect+0.25+0.25+0.75+0.75'

curl -d "figure     0.5    0.5  " \
	-X POST \
	http://localhost:17000/

curl -d "green  " \
	-X POST \
	http://localhost:17000/

curl 'http://localhost:17000/?cmd=figure+0.6+0.6'

curl -d "update" \
	-X POST \
	http://localhost:17000/