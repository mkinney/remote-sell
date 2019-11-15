version:
	git rev-parse HEAD > conf/git_version.txt

run: version
	revel run -v -a github.com/mkinney/remote-sell

package: version
	rm public/static/img/rs_*.png ; true
	GOOS=linux GOARCH=amd64 revel package -m prod github.com/mkinney/remote-sell

deploy:
	scp remote-sell.tar.gz bcnw_test:/tmp

deploy_prod:
	scp remote-sell.tar.gz bcnw:/tmp
