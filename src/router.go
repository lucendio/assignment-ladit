package main


import (
    "encoding/base64"
    "log"
    "net/http"
    "strings"

    "blocksvc/blocking"

    "github.com/gin-gonic/gin"
)


var blocklist = blocking.New()

var router = func() *gin.Engine {
    r := gin.New()
    r.Use( gin.Recovery() )
    // TODO: adjust logger according to configured log level and environment
    r.Use( gin.Logger() )

    blocklist.Add( "172.2.2.0/24", 3600 )

    // NOTE: including the health endpoint in the blocking
    //       middleware holds the risk of sb adding a CIDR
    //       that could exclude the client in the surrounding
    //       environment too and thus would render the service
    //       unhealthy from the outer perspective
    r.Any( "/healthcheck", Health )

    allowed := r.Group( "/", func( ctx *gin.Context ){
        clientIp := ctx.ClientIP()
        // TESTING
        // clientIp := ctx.GetHeader("X-Forwarded-For")
        if isBlocked, _ := blocklist.IsBlocked( clientIp ); isBlocked {
            ctx.Status( http.StatusForbidden )
            ctx.Abort()
        } else {
            ctx.Next()
        }
    })
    allowed.GET( "/stats", Statistics )

    restricted := allowed.Group( "/", func( ctx *gin.Context ){
        const AUTH_SCHEME = "Bearer"
        header := ctx.GetHeader( "Authorization" )
        if strings.HasPrefix( header, AUTH_SCHEME) == false {
            ctx.Status( http.StatusMethodNotAllowed )
            ctx.Abort()
            return
        }

        base64EncodedToken := strings.TrimSpace( header[ len(AUTH_SCHEME) : ] )
        base64DecodedToken, err := base64.StdEncoding.WithPadding(base64.NoPadding).DecodeString( base64EncodedToken )
        if err != nil {
            ctx.Status( http.StatusBadRequest )
            ctx.Abort()
            return
        }

        // NOTE: remove trailing newline(s) in case they were added during
        //       encoding (default behaviour)
        token := strings.TrimSpace( string( base64DecodedToken ) )
        if strings.Compare( token, DEFAULT_ACCESS_TOKEN ) != 0 {
            ctx.Status( http.StatusUnauthorized )
            ctx.Abort()
            return
        }

        ctx.Next()
    })
    restricted.POST( "/block", BlockCIDR )

    return r
}()



func Health( ctx *gin.Context ){
    ctx.String( http.StatusOK, "%s - Health", ctx.Request.Method )
}


func Statistics( ctx *gin.Context ){
    ctx.String( http.StatusOK, "%s - Statistics", ctx.Request.Method )
}


func BlockCIDR( ctx *gin.Context ){
    ctx.String( http.StatusOK, "%s - BlockCIDR", ctx.Request.Method )
}
