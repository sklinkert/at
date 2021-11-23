
PHONY: deploy
deploy:
	 aws s3 cp ../chart.html s3://at-data/chart.html
#	 aws s3 cp ../candles.csv s3://at-data/candles.csv
	 aws s3api put-object-acl --bucket at-data --key chart.html --acl public-read
#	 aws s3api put-object-acl --bucket at-data --key candles.csv --acl public-read
