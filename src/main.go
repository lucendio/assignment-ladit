package main

import (
    "blocksvc/configuration"
    "context"
    "errors"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
)


func main() {
    config, err := configuration.New()
    if err != nil {
        log.Fatalf( "HTTP server failed to start: %v", err )
    }

    router, _ := newRouter( config )

    server := &http.Server{
        Addr: fmt.Sprintf( "%s:%d", config.Host, config.Port ),
        Handler: router,
    }

    go func(){
        err := server.ListenAndServe()
        if err != nil {
            if errors.Is( err, http.ErrServerClosed ){
                log.Println( "HTTP server closed" )
            } else {
                log.Fatalf( "HTTP server failed to start: %v", err )
            }
        }
    }()

    osSignaling := make( chan os.Signal, 1 )
    signal.Notify( osSignaling, syscall.SIGHUP )
    signal.Notify( osSignaling, syscall.SIGINT )
    signal.Notify( osSignaling, syscall.SIGTERM )
    signal.Notify( osSignaling, syscall.SIGQUIT )

    shuttingDown := context.TODO()

    for {
        select {
            case <-osSignaling:
                log.Println( "Gracefully shutting down HTTP server" )

                var concludeShutdown context.CancelFunc
                shuttingDown, concludeShutdown = context.WithTimeout( context.Background(), time.Second * 15 )
                err := server.Shutdown( shuttingDown )
                if err != nil {
                    log.Printf("HTTP server failed to shut down: %v", err)
                }
                concludeShutdown()

            case <-shuttingDown.Done():
                log.Printf("Done and Done!")
                return
        }
    }
}
