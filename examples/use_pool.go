package main

import (
    "log"
    "sync/atomic"
    "fmt"
    "time"
    "math/rand"
    "sync"

    "../pool"
    "io"
)

const(
    maxGoroutineNum         = 20
    connectionPoolSize uint = 2
)

var idCounter int32

type dbConnection struct {
    ID int32
}

func (c *dbConnection) Close() error {
    log.Println("close db connection: ", c.ID)
    return nil
}

func createDbConnection() (io.Closer, error) {
    id := atomic.AddInt32(&idCounter, 1)
    log.Println("create db connection: ", id)
    return &dbConnection{
        ID: id,
    }, nil
}

func performQuries(query int, p *pool.Pool) error {
    if p == nil {
        return fmt.Errorf("pool is nil")
    }
    conn, err := p.Acquire()
    if err != nil {
        return err
    }

    defer p.Release(conn)

    time.Sleep(time.Duration(rand.Intn(1000))*time.Millisecond)
    log.Printf("QID[%d] CID[%d]\n", query, conn.(*dbConnection).ID)

    return nil
}

func main() {
    var wg sync.WaitGroup
    wg.Add(maxGoroutineNum)
    connPool, err := pool.New(createDbConnection, connectionPoolSize)
    if err != nil {
        log.Println(err)
        return
    }
    defer connPool.Close()

    for query := 0; query < maxGoroutineNum; query ++ {
        go func() {
            if err := performQuries(query, connPool); err != nil {
                log.Println(err)
            }
            wg.Done()
        }()
    }
    wg.Wait()
    log.Println("shutdone the program")
}
