# tsk

Tsk is a Go library for handling generic concurrent tasks.  

It provides multiple packages that each implement a single pattern.

`go get -u github.com/abevier/tsk`

## Packages
- [futures](#futures)
- [batch](#batch)
- [ratelimiter](#ratelimiter)
- [taskqueue](#taskqueue)

### futures
The futures package provides a wrapper around an asynchronous computation.  

A channel and a goroutine are often used to acheive the same result, but there are differences.  
- A Future can be read by any number of go routines, a channel only provides the value once
- A Future can be safely completed multiple times: only the first completion is used,
  subsequent completions are ignored silently, a channel will panic if written to when
  closed.

**Basic Future Example:**
```go
fut := futures.New[Response]()

go func() {
   bgCtx := context.WithTimeout(context.Background(), 45 * time.Second)
   resp, err := slowServiceClient.MakeRequest(bgCtx, req)
   if err != nil {
     fut.Fail(err)
     return
   }
   fut.Complete(resp)
}()

// This will block until the slowServiceClient returns a response and the future is completed
resp, err := fut.Get(ctx)
```

**Another basic Future example using FromFunc**
```go
fut := futures.FromFunc(func() (Response, error) {
   bgCtx := context.WithTimeout(context.Background(), 45 * time.Second)
   resp, err := slowServiceClient.MakeRequest(bgCtx, req)
   return resp, err
})

// This will block until the slowServiceClient returns a response and the future is completed
resp, err := fut.Get(ctx)
```

**Simple Cache using Future Example**
```go
type SimpleCache struct {
  m sync.Mutex{}
  values map[string]*futures.Future[Response]
}

func NewCache() *SimpleCache {
  return &SimpleCache{
    m: sync.Mutex{},
    values: make(map[string]*futures.Future[Response])

  }
}

// Get will return a future if it exists, if it does not then a new future will
// be returned after it is added to the cache
func (c *SimpleCache) Get(key string) *futures.Future[Response] {
  c.m.Lock()
  defer c.m.Unlock()

  f, ok := c.values[key]
  if !ok {
    f = futures.FromFunc(func() (Response, error) {
      bgCtx := context.WithTimeout(context.Background(), 45 * time.Second)
      resp, err := slowServiceClient.MakeRequest(bgCtx, Request{Key: key})
      return resp, err
    })

    c.values[key] = f
  }
  return f
}

func main() {
  wg := sync.WaitGroup{}

  cache := NewCache()

  // All go routines read the same result - only one slow request is made despite 100 
  // concurrent requests for the same key
  for i := 0; i < 100; i++ {
    wg.Add(1)
    go func() {
      defer wg.Done()
      futureResp := cache.Get("the-key")
      resp, _ := futureResp.Get(context.TODO())
      log.Printf("response is: %v", resp)
    }()
  }

  wg.Wait()
}
```

### batch
The batch package provides a batch executor implemenation that allows multiple go routines to seamlessly batch tasks
which are then flushed to a user defined executor function. 

A common use case for this is batching multiple http requests that wish to write to a database into a single update 
to the database.  Another common use is multiple http requests that should publish a message to AWS SQS which allows 
for batching of up to 10 messages in a single request.

**Batch Example**
```go
be := batch.New(batch.Opts{MaxSize: 10, MaxLinger:250 * time.Millisecond}, runBatch)

func runBatch(events []string) ([]results.Result[string], error) {
  serviceRequest := myservice.Request{items: events}
  resp, err := myservice.Write(serviceRequest)
  if err != nil {
    // All items in the batch will fail with this error
    return nil, err
  }

  // The run batch function must return an result for every item passed into the function
  var res []results.Result[string]
  for _, r := range resp.items {
    if r.Err != nil {
      res = append(res, results.Fail[string](err))
      continue
    }
    res = append(res, results.Success(item.ID))
  }

  return res, nil
}

func eventHandler(w http.ResponseWriter, req *http.Request) {
  b, _ := ioutil.ReadAll(req.body)
  event := string(b)

  // Each request will submit individual items to the batch which will flush after 10 items or once the batch is 250ms old
  // This will not return until the batch flushes and a result is returned from runBatch
  id, err := be.Submit(req.Context(), event)
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    return 
  }
  w.Write([]byte(id))
}

func main() {
  http.HandleFunc("/event", eventHandler)
  http.ListenAndServe(":8080", nil)
}
```


### ratelimiter
The ratelimiter package provides a rate limiter that utilizes the Token Bucket algorithm to limit the rate that
a function will be invoked.  A common use case is to prevent a service from overwhelming other systems.

**Rate Limiter Example**
```go
opts := ratelimiter.Opts{
  Limit: 10,
  Burst: 1,
  MaxQueueDepth: 100,
  FullQueueStategy: ratelimiter.ErrorWhenFull,
}
rl := ratelimiter.New(opts, do)

// This function will be called up to 10 times per second
func do(ctx context.Context, request string) (string, error) {
  resp, err := myservice.RateLimitedQuery(ctx, request)
  return resp, err
}

func requestHandler(w http.ResponseWriter, req *http.Request) {
  b, _ := ioutil.ReadAll(req.body)
  request := string(b)

  // Each call to submit will block until the rate limited function is invoked and returns a result.
  // If the number pending request is too high ErrQueueFull is returned and the server will respond 
  // to the request with an HTTP 429
  resp, err := rl.Submit(req.Context(), event)
  if err != nil {
    if errors.Is(err, ratelimiter.ErrQueueFull) {
      w.WriteHeader(http.StatusTooManyRequests)
    } else {
      w.WriteHeader(http.StatusInternalServerError)
    }
    return 
  }
  w.Write([]byte(resp))
}

func main() {
  http.HandleFunc("/request", requestHandler)
  http.ListenAndServe(":8080", nil)
}
```

### taskqueue
The taskqueue package provides a task queue that schedules work on a pool of goroutines. It is used to limit concurrent
invocations to a function.  Much like a rate limiter a common use case is to prevent overwhelming other services.

**Task Queue Example**
```go
opts := taskqueue.Opts{
  MaxWorkers: 3,
  MaxQueueDepth: 100,
  FullQueueStategy: taskqueue.ErrorWhenFull,
}
tq := taskqueue.New(opts, do)

// This function will never be invoked more than 3 times concurrently while behind the TaskQueue
func do(ctx context.Context, request string) (string, error) {
  resp, err := myservice.ConcurrencyLimitedQuery(ctx, request)
  return resp, err
}

func requestHandler(w http.ResponseWriter, req *http.Request) {
  b, _ := ioutil.ReadAll(req.body)
  request := string(b)

  // Each call to submit will block until the concurrency limited function is invoked and returns a result.
  // If the number pending request is too high ErrQueueFull is returned and the server will respond 
  // to the request with an HTTP 429
  resp, err := tq.Submit(req.Context(), event)
  if err != nil {
    if errors.Is(err, taskqueue.ErrQueueFull) {
      w.WriteHeader(http.StatusTooManyRequests)
    } else {
      w.WriteHeader(http.StatusInternalServerError)
    }
    return 
  }
  w.Write([]byte(resp))
}

func main() {
  http.HandleFunc("/request", requestHandler)
  http.ListenAndServe(":8080", nil)
}
```

## Motivation
I wrote this library because I found that I was repeating these patterns in multiple places.  My preference is to
keep application business logic and task plumbing logic seperate when at all possible.  This means I try to avoid
exposing channels and other concurreny primitives to business services.  Many concurrency bugs I've encountered in 
Go have been due to improperly sending/receiving to unbuffered channels and forgetting to cancel when a context is
cancelled. 
