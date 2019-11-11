# Assume that went thru the revel getting started (go installed, revel installed)

go get -u github.com/skip2/go-qrcode/...

# generated with
revel new github.com/mkinney/remote-sell

# do development like this
revel run -a  github.com/mkinney/remote-sell

Notes:
- Needed to deal with insecure https since "x509: certificate is valid for master.batm.generalbytes.com, not localhost". 
  This is a security risk.

# TODO:
- When/how to clean up old qr code tmp files?
- Service


