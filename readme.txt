# Assume that went thru the revel getting started (go installed, revel installed)

go get -u github.com/skip2/go-qrcode/...

# generated with
revel new github.com/mkinney/remote-sell

# do development like this
revel run -a  github.com/mkinney/remote-sell

# To create/start service: (must be root)
# copy over the remote_sell.service file to /home/rs/remote_sell.service
cp /home/rs/remote_sell.service /etc/systemd/system/remote_sell.service
systemctl daemon-reload
systemctl start remote_sell.service
systemctl status remote_sell.service
systemctl enable remote_sell.service

Notes:
- Needed to deal with insecure https since "x509: certificate is valid for master.batm.generalbytes.com, not localhost". 
  This is a security risk.
- Had to run this:
  ln -s src/github.com/mkinney/remote-sell/public/ public

# TODO:
- When/how to clean up old qr code tmp files?


