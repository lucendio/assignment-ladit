package main

import (
    "blocksvc/blocking"
    "blocksvc/configuration"
    "encoding/base64"
    "fmt"
    "net/http"
    "net/http/httptest"
    "os"
    "strings"
    "testing"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
)


func setup() (*gin.Engine, *blocking.Blocklist) {
    if os.Getenv( "ACCESS_TOKEN" ) == "" {
        os.Setenv( "ACCESS_TOKEN", "some-token" )
    }

    config, _ := configuration.New()
    return newRouter( config )
}


func TestBlockingMiddleware( t *testing.T ){
    const ttl = 2
    const clientIp = "10.0.0.1"

    router, blocklist := setup()

    // NOTE: using an incoming server request to mimic client IP
    req := httptest.NewRequest( "GET", "/stats", nil )
    req.RemoteAddr = fmt.Sprintf( "%s:%d", clientIp, 1234 )

    // NOTE: immediately expires and is being removed from the block list
    blocklist.Add( "10.0.0.0/24", int32( 0 ) )

    res := httptest.NewRecorder()
    router.ServeHTTP( res, req )
    assert.Equal( t, http.StatusOK, res.Code )

    blocklist.Add( "10.0.0.0/24", int32( ttl ) )

    res = httptest.NewRecorder()
    router.ServeHTTP( res, req )
    assert.Equal( t, http.StatusForbidden, res.Code )

    time.Sleep( time.Second * time.Duration( ttl ) )

    res = httptest.NewRecorder()
    router.ServeHTTP( res, req )
    assert.Equal( t, http.StatusOK, res.Code )
}


func TestAuthenticationMiddleware( t *testing.T ){
   const wrongToken = "wrong"
   wrongTokenEncoded := base64.StdEncoding.EncodeToString( []byte( wrongToken ) )
   const rightToken = "right"
   rightTokenEncoded := base64.StdEncoding.EncodeToString( []byte( rightToken ) )
   const invalidEncodedBarerToken = "-"
   os.Setenv( "ACCESS_TOKEN", rightToken )

   router, _ := setup()

   req := httptest.NewRequest( "POST", "/block", nil )
   res := httptest.NewRecorder()

   // NOTE: no auth header
   router.ServeHTTP( res, req )
   assert.Equal( t, http.StatusMethodNotAllowed, res.Code )

   req.Header.Set( "Authorization", fmt.Sprintf( "%s %s", "Foo", wrongTokenEncoded ) )
   router.ServeHTTP( res, req )
   assert.Equal( t, http.StatusMethodNotAllowed, res.Code )

   req.Header.Set( "Authorization", fmt.Sprintf( "%s %s", "Bearer", invalidEncodedBarerToken ) )
   res = httptest.NewRecorder()
   router.ServeHTTP( res, req )
   assert.Equal( t, http.StatusBadRequest, res.Code )

   req.Header.Set( "Authorization", fmt.Sprintf( "%s %s", "Bearer", wrongTokenEncoded ) )
   res = httptest.NewRecorder()
   router.ServeHTTP( res, req )
   assert.Equal( t, http.StatusUnauthorized, res.Code )

   req.Header.Set( "Authorization", fmt.Sprintf( "%s %s", "Bearer", rightTokenEncoded ) )
   res = httptest.NewRecorder()
   router.ServeHTTP( res, req )
   assert.Equal( t, http.StatusBadRequest, res.Code )
}


func TestHealthHandler( t *testing.T ){
   router, _ := setup()

   res := httptest.NewRecorder()
   req, _ := http.NewRequest( "GET", "/healthcheck", nil )
   router.ServeHTTP( res, req )

   assert.Equal( t, http.StatusOK, res.Code )
}


func TestStatisticsHandler( t *testing.T ){
    const ttl = 3600
    const blockedClientIp = "10.0.0.1"
    const allowedClientIp = "192.168.0.1"

    router, blocklist := setup()

    req := httptest.NewRequest( "GET", "/stats", nil )
    req.RemoteAddr = fmt.Sprintf( "%s:%d", allowedClientIp, 1234 )

    res := httptest.NewRecorder()
    router.ServeHTTP( res, req )
    assert.Equal( t, http.StatusOK, res.Code )
    assert.JSONEq( t, `{"cidrs": 0, "blocked_requests": 0, "accepted_requests": 1}`, res.Body.String() )

    blocklist.Add( "10.0.0.0/24", int32( ttl ) )

    res = httptest.NewRecorder()
    router.ServeHTTP( res, req )
    assert.Equal( t, http.StatusOK, res.Code )
    assert.JSONEq( t, `{"cidrs": 1, "blocked_requests": 0, "accepted_requests": 2}`, res.Body.String() )

    req.RemoteAddr = fmt.Sprintf( "%s:%d", blockedClientIp, 1234 )
    router.ServeHTTP( httptest.NewRecorder(), req )

    req.RemoteAddr = fmt.Sprintf( "%s:%d", allowedClientIp, 1234 )
    res = httptest.NewRecorder()
    router.ServeHTTP( res, req )
    assert.Equal( t, http.StatusOK, res.Code )
    assert.JSONEq( t, `{"cidrs": 1, "blocked_requests": 1, "accepted_requests": 3}`, res.Body.String() )
}


func TestBlockCidrHandler( t *testing.T ){
    const wrongBodyInvalidJSON = `{ "cidr: , ttl }`
    const wrongBodyInvalidValue = `{ "cidr": "10.", "ttl": 1 }`
    const wrongBodyMissingAttribute = `{ "cidr": "10.0.0.0/24" }`
    const rightBody = `{ "cidr": "10.0.0.0/24", "ttl": 600 }`
    const rightBodySameCidr = `{ "cidr": "10.0.0.0/24", "ttl": 3600 }`
    const allowedClientIp = "192.168.0.1"

    const rightToken = "right"
    rightTokenEncoded := base64.StdEncoding.EncodeToString( []byte( rightToken ) )
    os.Setenv( "ACCESS_TOKEN", rightToken )

    router, _ := setup()

    run := func( body string ) *httptest.ResponseRecorder {
        req := httptest.NewRequest( "POST", "/block", strings.NewReader( body ) )
        req.RemoteAddr = fmt.Sprintf( "%s:%d", allowedClientIp, 1234 )
        req.Header.Set( "Authorization", fmt.Sprintf( "%s %s", "Bearer", rightTokenEncoded ) )
        res := httptest.NewRecorder()
        router.ServeHTTP( res, req )
        return res
    }

    res := run( wrongBodyInvalidJSON )
    assert.Equal( t, http.StatusBadRequest, res.Code )

    res = run( wrongBodyInvalidValue )
    assert.Equal( t, http.StatusBadRequest, res.Code )

    res = run( wrongBodyMissingAttribute )
    assert.Equal( t, http.StatusBadRequest, res.Code )

    res = run( rightBody )
    assert.Equal( t, http.StatusOK, res.Code )

    res = run( rightBodySameCidr )
    assert.Equal( t, http.StatusForbidden, res.Code )
}
