## Dependency

* [gin][1] as web framework
* [go-mailer][2] as mailer with a web api
* [bolt][3] store database
* [jwt-go][4] authentication
* [kinpin][5] command-line option parser

### install

    get get -u github.com/fengxsong/pubmgmt

### get start

    cd $GOPATH/src/github.com/fengxsong/pubmgmt
    go run app/pubmgmt.go --help

NOW IT'S IN EARLY STATE(backend is almost done) BUT... FELL FREE TO HAVE A TRY :)

I had only code `subversion` and a very simple `shell` module.

YOU CAN WRITE MODULES BY YOUR OWN. JUST IMPLEMENT THE `Module interface`


ATTENTION: In this project, I draw on some idea from [portainer][6]


[1]: https://gin-gonic.github.io/gin/
[2]: https://github.com/kataras/go-mailer
[3]: https://github.com/boltdb/bolt
[4]: https://github.com/dgrijalva/jwt-go
[5]: https://gopkg.in/alecthomas/kingpin.v2
[6]: https://github.com/portainer/portainer