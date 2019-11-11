run:
	revel run -a github.com/mkinney/remote-sell

package:
	GOOS=linux GOARCH=amd64 revel package -m prod github.com/mkinney/remote-sell

deploy:
	scp remote-sell.tar.gz bcnw_test:/tmp
