package main

import (
    //"fmt"
	_ "math"
	_ "testing"
    "fmt"
    "bytes"
    "io"
	"time"
    "github.com/juju/ratelimit"
)


type reader struct{
    r io.Reader
    tbLimiter *ratelimit.Bucket
}

func NewReader(r io.Reader, tb *ratelimit.Bucket) io.Reader {
    return &reader{
        r: r,
        tbLimiter: tb,
    }
}


func (r *reader) Read(buf []byte) (int, error){
    //fmt.Println("Read buf size: ",int64(len(buf)))
    //waitTime := r.tbLimiter.Take(int64(len(buf)))
	_, waitTime := r.tbLimiter.TakeMaxDuration(int64(len(buf)), 0)
    if !waitTime{
        //time.Sleep(waitTime)
		return 0, nil
    }
    n, err := r.r.Read(buf)
    if n <= 0 {
        return n, err
    }
    //fmt.Printf("Write bytes : %d\n", n)
    return n, err
}


func main(){
    counter := 0
    timeslot := 0
    src := bytes.NewReader(make([]byte, 1024*1024))
    dst := &bytes.Buffer{}
    start := time.Now()

    tbLimiter := ratelimit.NewBucketWithQuantum(time.Second, 100*1024, 100*1024)
    r := NewReader(src, tbLimiter)
    rateChecker := time.NewTicker(time.Second * 1)
    buf := make([]byte, 10*1024)
L1:
    for{
        select{
        case <- rateChecker.C:
            fmt.Printf("time slot %d : rate is %d\n", timeslot, counter)
            timeslot++
            counter = 0
        default:
            if n, err := r.Read(buf); err == nil {
                dst.Write(buf[0:n])
                if n != 0{
                    counter += n
                }
            }else{
                break L1
            }
        }
    }
    fmt.Printf("Copied %d bytes in %s\n", dst.Len(), time.Since(start))

}
