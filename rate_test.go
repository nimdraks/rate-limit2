package main

import(
    "io"
    "testing"
    "github.com/juju/ratelimit"
    "net/http"
    "net/http/httptest"
    "github.com/stretchr/testify/assert"
    "sync"
    "fmt"
    "time"
)


func BenchmarkMakeLimiter(b *testing.B){
    for i:=1; i< b.N; i++{
        ratelimit.NewBucketWithQuantum(time.Second, int64(i), int64(i))
    }
}

func BenchmarkUsingLimiter(b *testing.B){
    as := assert.New(b)
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
        if _, err := io.WriteString(w, "I love you"); err != nil{
            b.Fail()
        }
    }))
    defer ts.Close()

    ServerURL := ts.URL
    httpClient := http.Client{}


    for i:=1; i< b.N; i++{
        Limiter := ratelimit.NewBucketWithQuantum(time.Second, int64(i), int64(i))
        var wait sync.WaitGroup
        wait.Add(i)
        for j:=1; j <= i; j++{
            go func(){
                defer wait.Done()
                _, resultBool := Limiter.TakeMaxDuration(1, 0)
                body, err := httpClient.Get(ServerURL)
                body.Body.Close()
                as.Equal(true, resultBool)
                as.Equal(nil, err)
            }()
        }
        wait.Wait()
    }
}

func TestUsingLimiterScale(t *testing.T){
    start := time.Now()

    testSize := 200
    as := assert.New(t)
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
        if _, err := io.WriteString(w, "I love you"); err != nil{
            t.Fail()
        }
    }))
    defer ts.Close()

    ServerURL := ts.URL
    httpClient := http.Client{}

    var wait sync.WaitGroup
    wait.Add(testSize)
    Limiter := ratelimit.NewBucketWithQuantum(time.Second, int64(testSize), int64(testSize))

    duration := time.Since(start)
    fmt.Println(duration)

    start = time.Now()
    for i:=1; i<= testSize; i++{
        go func(){
            defer wait.Done()
            _, resultBool := Limiter.TakeMaxDuration(1, 0)
            body, err := httpClient.Get(ServerURL)
            body.Body.Close()
            as.Equal(true, resultBool)
            as.Equal(nil, err)
        }()
    }

    wait.Wait()

    duration = time.Since(start)
    fmt.Println(duration)

}
