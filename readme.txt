# Assume that went thru the revel getting started (go installed, revel installed)

go get -u github.com/skip2/go-qrcode/...

# generated with
revel new github.com/mkinney/remote-sell

# do development like this
revel run -a  github.com/mkinney/remote-sell

# TODO:
- When/how to clean up old qr code tmp files?
- Need to deal with insecure https?
- Makefile for building/deployment
- Service


