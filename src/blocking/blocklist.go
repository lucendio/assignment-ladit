package blocking

import (
    "errors"
    "net"
    "time"
)


type blockentry struct {
    cidr net.IPNet
    ttl int32
    createdAt time.Time
}


func (be *blockentry) hasExpired() bool {
    now := time.Now()
    expiresAt := be.createdAt.Add( time.Second * time.Duration( be.ttl ) )
    return now.After( expiresAt )
}



type Blocklist struct {
    entries map[string]*blockentry
    accepted int
    blocked int
}


func New() *Blocklist {
    return &Blocklist{
        entries: make( map[string]*blockentry ),
        accepted: 0,
        blocked: 0,
    }
}


func (bl *Blocklist) Add( cidr string, ttl int32 ) error {
    _, ipNet, err := net.ParseCIDR( cidr )
    if err != nil {
        return err
    }

    if _, exists := bl.entries[ cidr ]; exists {
        return errors.New( "blocked CIDR cannot be overwritten" )
    }

    bl.entries[ipNet.String()] = &blockentry{
        cidr: *ipNet,
        ttl: ttl,
        createdAt: time.Now(),
    }

    return nil
}


func (bl *Blocklist) IsBlocked( ipAddress string ) (bool, error) {
    ip := net.ParseIP( ipAddress )
    if ip == nil {
        return true, errors.New( "invalid IP address" )
    }

    for identifier, entry := range bl.entries {
        if entry.cidr.Contains( ip ) {
            if entry.hasExpired() {
                // NOTE: parallel execution due to multiple incoming requests
                //       at the same time won't be an issue here, since
                //       'delete' is a no-op, if entry doesn't exist anymore
                defer delete( bl.entries, identifier )

                // NOTE: due to the entry being expired, it won't count as
                //       blocked, but there might still be another CIDR in
                //       the blocklist that contains the given IP
                continue
            }

            bl.blocked += 1
            return true, nil
        }
    }

    bl.accepted += 1
    return false, nil
}


type statistics struct {
    Amount int      `json:"cidrs"`
    Blocked int     `json:"blocked_requests"`
    Accepted int    `json:"accepted_requests"`
}

func (bl *Blocklist) Statistics() *statistics {
    return &statistics{
        Amount: len(bl.entries),
        Accepted: bl.accepted,
        Blocked: bl.blocked,
    }
}
