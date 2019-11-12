# remote-sell

Repo for doing remote sell processing of crypto from Bitcoin ATMs.

To get this working, I assume that you went thru the revel getting started (go installed, revel installed)

    go get -u github.com/skip2/go-qrcode/...

I initially generated with this command:

    revel new github.com/mkinney/remote-sell

Do development like this:

    revel run -a  github.com/mkinney/remote-sell

or

    make run

To create/start service: (must be root)

    # copy over the remote_sell.service file to /home/rs/remote_sell.service
    # cp /home/rs/remote_sell.service /etc/systemd/system/remote_sell.service
    # systemctl daemon-reload
    # systemctl start remote_sell.service
    # systemctl status remote_sell.service
    # systemctl enable remote_sell.service

To deploy

    make package
    make deploy

On remote:

    systemctl stop remote_sell ; cd /home/rs/rs/ ; tar zxf /tmp/remote-sell.tar.gz ; rm -rf public/img/rs*png ; systemctl start remote_sell

You probably want to change batm_url values in conf/app.conf.

If you add a new BATM (or crypto), need to add it to few places:
1) in app/controllers/app.go's LocationToSerialNumber function
2) may need to add new crypto prefix to CryptoToPrefix
3) in app/views/App/Index.html in the 'select' dropdown

# Notes:
* Needed to deal with insecure https since "x509: certificate is valid for master.batm.generalbytes.com, not localhost".
  This is a security risk.

* Ran these commands:

    ln -s src/github.com/mkinney/remote-sell/public/ public
    ln -s src/github.com/mkinney/remote-sell/log/ log
    ln -s src/github.com/mkinney/remote-sell/conf/ conf

# TODO:
* When/how to clean up old qr code tmp files?

# Welcome to Revel

A high-productivity web framework for the [Go language](http://www.golang.org/).


### Start the web server:

   revel run myapp

### Go to http://localhost:9000/ and you'll see:

    "It works"

## Code Layout

The directory structure of a generated Revel application:

    conf/             Configuration directory
        app.conf      Main app configuration file
        routes        Routes definition file

    app/              App sources
        init.go       Interceptor registration
        controllers/  App controllers go here
        views/        Templates directory

    messages/         Message files

    public/           Public static assets
        css/          CSS files
        js/           Javascript files
        images/       Image files

    tests/            Test suites


## Help

* The [Getting Started with Revel](http://revel.github.io/tutorial/gettingstarted.html).
* The [Revel guides](http://revel.github.io/manual/index.html).
* The [Revel sample apps](http://revel.github.io/examples/index.html).
* The [API documentation](https://godoc.org/github.com/revel/revel).

