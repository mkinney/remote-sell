run:
	revel run -v -a github.com/mkinney/remote-sell

package:
	rm public/static/img/rs_*.png ; true
	GOOS=linux GOARCH=amd64 revel package -m prod github.com/mkinney/remote-sell

deploy:
	scp remote-sell.tar.gz bcnw_test:/tmp
