# remote-sell

Repository for doing remote sell processing of crypto from Bitcoin ATMs. A "remote sell" is when a user does not need to be next to the BATM for a request to sell some crypto. They can visit this page (most likely from their cellphone), enter the amount of fiat and crypto to sell. A QR code will be generated that they can use to send crypto. Once the crypto confirms, they can go to the BATM and get their money using the QR code. This service is needed because some cryptos (like BTC) can take more than 10 minutes to confirm. So, a person may not want to wait around the BATM for confirmations. They can "plan ahead".

# BATM setup
For BATM side, you will need to clone/build https://github.com/GENERALBYTESCOM/batm_public. Be sure to uncomment `server_extensions_extra/src/main/resources/batm-extensions.xml` then follow the build steps. For instance, the line should look like this:

    <extension class="com.generalbytes.batm.server.extensions.extra.examples.rest.RESTExampleExtension" />

Build the new jar. Copy over the jar to temp location on the BATM server. Stop BATM (`batm-manage stop all`). Copy the jar file into correct location (`cp /tmp/new.jar /batm/app/master/extensions/batm_server_extensions_extra.jar`. Start BATM. Verify that you can hit basic endpoint.

# Development of remote_sell service
For this service, I am development on a mac and code runs on Ubuntu.

To do remote_sell development, I assume that you went thru the revel getting started. (go installed, revel installed)

    go get -u github.com/skip2/go-qrcode/...

I initially generated this remote_sell service with this command:

    revel new github.com/mkinney/remote-sell

Do development like this:

    revel run -a  github.com/mkinney/remote-sell

or simply:

    make run

# To prepare for initial deployment of remote_sell service:

To create/start service: (must be root)

    # copy over the remote_sell.service file from repo to /home/rs/remote_sell.service
    # (do the next steps as root)
    # cp /home/rs/remote_sell.service /etc/systemd/system/remote_sell.service
    # systemctl daemon-reload
    # systemctl start remote_sell.service
    # systemctl status remote_sell.service
    # systemctl enable remote_sell.service

* Run these commands (only to run these once as `rs` user):

    ln -s src/github.com/mkinney/remote-sell/public/ public
    ln -s src/github.com/mkinney/remote-sell/log/ log
    ln -s src/github.com/mkinney/remote-sell/conf/ conf

# Deploy a new version (from mac):

    make package
    make deploy

On remote:

    systemctl stop remote_sell ; cd /home/rs/rs/ ; tar zxf /tmp/remote-sell.tar.gz ; rm -rf public/img/rs*png ; systemctl start remote_sell

To make this easier, in the root/.bashrc setup this alias

    alias drs='systemctl stop remote_sell ; cd /home/rs/rs/ ; tar zxf /tmp/remote-sell.tar.gz ; rm -rf public/img/rs*png ; systemctl start remote_sell'

That way you can just run `drs` to do all of that on the server side.

# Notes:
* For local testing, you may want to clone/run the https://github.com/mkinney/simulate-batm repo.

* You probably want to change batm_url values in conf/app.conf.

* If you add a new BATM (or crypto), need to add it to few places:
1) in app/controllers/app.go's LocationToSerialNumber function
2) may need to add new crypto prefix to CryptoToPrefix
3) in app/views/App/Index.html in the 'select' dropdown

* To see build version on running instance: (be sure to have terminal width wide enough)

    journalctl -u remote_sell

or

    systemctl status remote_sell

* For development, revel logging wants an odd number of args. So, typically do something like this:

    c.Log.Info("in getCryptoAmount", "full_url", full_url)

* Needed to deal with insecure https since "x509: certificate is valid for master.batm.generalbytes.com, not localhost".
  This is a security risk.


# TODO:
* When/how to clean up old qr code tmp files?

# Code Layout

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

# Help

* The [Getting Started with Revel](http://revel.github.io/tutorial/gettingstarted.html).
* The [Revel guides](http://revel.github.io/manual/index.html).
* The [Revel sample apps](http://revel.github.io/examples/index.html).
* The [API documentation](https://godoc.org/github.com/revel/revel).
