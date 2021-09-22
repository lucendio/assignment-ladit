package main

import (
    "blocksvc/blocking"
    "context"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
)



const (
    DEFAULT_PORT = 3000
    DEFAULT_HOST = "localhost"
)



func main() {
    osSignaling := make( chan os.Signal, 1 )
    signal.Notify( osSignaling, syscall.SIGHUP )
    signal.Notify( osSignaling, syscall.SIGINT )
    signal.Notify( osSignaling, syscall.SIGTERM )
    signal.Notify( osSignaling, syscall.SIGQUIT )

    router := http.NewServeMux()
    router.HandleFunc("/healthcheck", func( res http.ResponseWriter, req *http.Request ){
        io.WriteString( res, "Hello, world!\n" )
    })

    server := &http.Server{
        Addr: fmt.Sprintf( "%s:%d", DEFAULT_HOST, DEFAULT_PORT ),
        Handler: router,
    }

    go func(){
        err := server.ListenAndServe()
        if err != nil {
            if err == http.ErrServerClosed {
                log.Println( "HTTP server closed" )
            } else {
                log.Fatalf( "HTTP server failed to start: %s", err )
            }
        }
    }()

    blocklist := blocking.New()
    blocklist.Add("192.168.2.0/24", 600)

    shuttingDown := context.TODO()

    for {
        select {
            case <-osSignaling:
                log.Println( "Gracefully shutting down HTTP server" )

                var finishShutdown context.CancelFunc
                shuttingDown, finishShutdown = context.WithTimeout( context.Background(), time.Second * 15 )
                err := server.Shutdown( shuttingDown )
                if err != nil {
                    log.Printf("HTTP server failed to shut down: %v", err)
                }
                finishShutdown()

            case <-shuttingDown.Done():
                log.Printf("Done and Done!")
                return
        }
    }
}
