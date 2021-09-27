package main

import (
    "blocksvc/configuration"
    "encoding/base64"
    "errors"
    "net"
    "net/http"
    "strings"

    "blocksvc/blocking"

    "github.com/gin-gonic/gin"
)


func newRouter( config *configuration.Config ) (*gin.Engine, *blocking.Blocklist) {
    router := gin.New()

    router.Use( gin.Recovery() )
    // TODO: adjust logger according to configured log level and environment
    router.Use( gin.Logger() )

    blocklist := blocking.New()

    // NOTE: including the health endpoint in the blocking middleware holds
    //       the risk of sb adding a CIDR that could also exclude the client
    //       responsible for checking the health and thus render the service
    //       unhealthy from its perspective
    router.Any( "/healthcheck", healthHandler )

    allowed := router.Group( "/", getBlockingMiddleware( blocklist ))
    allowed.GET( "/stats", getStatisticsHandler( blocklist ) )

    restricted := allowed.Group( "/", authenticationMiddleware( config ) )
    restricted.POST( "/block", getBlockCidrHandler( blocklist ) )

    return router, blocklist
}



func healthHandler( ctx *gin.Context ){
    // TODO: add a global healthiness state (bool) variable and serve respective code here
    ctx.AbortWithStatus( http.StatusOK )
}


func getStatisticsHandler( bl *blocking.Blocklist ) func( *gin.Context ) {
    return func( ctx *gin.Context ){
        ctx.JSON( http.StatusOK, bl.Statistics() )
    }
}


func getBlockCidrHandler( bl *blocking.Blocklist ) func( *gin.Context ) {
    type blockRequestBody struct {
        Cidr string `form:"cidr" json:"cidr" binding:"required"`
        Ttl int32   `form:"ttl" json:"ttl" binding:"required"`
    }

    return func( ctx *gin.Context ){
        var body blockRequestBody
        if err := ctx.ShouldBindJSON( &body ); err != nil {
            ctx.AbortWithStatus( http.StatusBadRequest )
            return
        }

        err := bl.Add( body.Cidr, body.Ttl )
        if err != nil {
            var cidrParseError *net.ParseError
            if errors.As( err, &cidrParseError ){
                ctx.AbortWithStatus( http.StatusBadRequest )
                return
            }
            if errors.Is( err, blocking.ErrCidrAlreadyExists ){
                // FUTUREWORK: it might be desired to inform the client about
                //             the TTL of the existing entry
                ctx.AbortWithStatus( http.StatusForbidden )
                return
            }
            // NOTE: undesired, no other option left at this point
            ctx.AbortWithStatus( http.StatusInternalServerError )
            return
        }

        ctx.AbortWithStatus( http.StatusOK )
    }
}


func getBlockingMiddleware( bl *blocking.Blocklist ) func( *gin.Context ) {
    return func( ctx *gin.Context ){
        clientIp := ctx.ClientIP()
        if isBlocked, _ := bl.IsBlocked( clientIp ); isBlocked {
            ctx.AbortWithStatus( http.StatusForbidden )
        } else {
            ctx.Next()
        }
    }
}


func authenticationMiddleware( config *configuration.Config ) func( *gin.Context ) {
    return func( ctx *gin.Context ){
        const AUTH_SCHEME = "Bearer"

        header := ctx.GetHeader( "Authorization" )
        if strings.HasPrefix( header, AUTH_SCHEME ) == false {
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
        if strings.Compare( token, config.AccessToken ) != 0 {
            ctx.Status( http.StatusUnauthorized )
            ctx.Abort()
            return
        }

        ctx.Next()
    }
}
