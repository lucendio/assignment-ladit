package main


import (
    "encoding/base64"
    "net/http"
    "strings"

    "blocksvc/blocking"

    "github.com/gin-gonic/gin"
)


var router = func() *gin.Engine {
    blocklist := blocking.New()

    r := gin.New()
    r.Use( gin.Recovery() )
    // TODO: adjust logger according to configured log level and environment
    r.Use( gin.Logger() )

    // TODO: remove
    blocklist.Add( "172.2.2.0/24", 3600 )

    // NOTE: including the health endpoint in the blocking middleware holds
    //       the risk of sb adding a CIDR that could also exclude the client
    //       responsible for checking the health and thus render the service
    //       unhealthy from its perspective
    r.Any( "/healthcheck", healthHandler )

    allowed := r.Group( "/", getBlockingMiddleware( blocklist ))
    allowed.GET( "/stats", getStatisticsHandler( blocklist ) )

    restricted := allowed.Group( "/", authenticationMiddleware )
    restricted.POST( "/block", getBlockCidrHandler )

    return r
}()



func healthHandler( ctx *gin.Context ){
    ctx.AbortWithStatus( http.StatusOK )
}


func getStatisticsHandler( bl *blocking.Blocklist ) func( *gin.Context ) {
    return func( ctx *gin.Context ){
        ctx.JSON( http.StatusOK, bl.Statistics() )
    }
}


func getBlockCidrHandler( ctx *gin.Context ){
    ctx.String( http.StatusOK, "%s - BlockCIDR", ctx.Request.Method )
}



func getBlockingMiddleware( bl *blocking.Blocklist ) func( *gin.Context ) {
    return func( ctx *gin.Context ){
        // TODO: for testing mock method with ctx.GetHeader("X-Forwarded-For")
        clientIp := ctx.ClientIP()
        if isBlocked, _ := bl.IsBlocked( clientIp ); isBlocked {
            ctx.AbortWithStatus( http.StatusForbidden )
        } else {
            ctx.Next()
        }
    }
}


func authenticationMiddleware( ctx *gin.Context ){
    const AUTH_SCHEME = "Bearer"

    header := ctx.GetHeader( "Authorization" )
    if strings.HasPrefix( header, AUTH_SCHEME) == false {
        ctx.Status( http.StatusMethodNotAllowed )
        ctx.Abort()
        return
    }

    base64EncodedToken := strings.TrimSpace( header[ len(AUTH_SCHEME) : ] )
    base64DecodedToken, err := base64.StdEncoding.DecodeString( base64EncodedToken )
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
}
