# CloseWaiter

## Why?
The standard sync.WaitGroup is not sufficient for the use case of closing a channel with multiple writers.

WaitGroup.Add and WaitGroup.Race must be synchronized with a mutex in order to ensure deterministic behavior.
Unfortunately during a shutdown, calling Add while Wait is holding the lock will cause the caller of Add to block
until the shutdown completes.  This implementation signals an error to the writer immediately instead of blocking.

See:
- https://www.leolara.me/blog/closing_a_go_channel_written_by_several_goroutines/
- https://groups.google.com/forum/#!topic/golang-nuts/Qq_h0_M51YM
- https://stackoverflow.com/questions/53769216/is-it-safe-to-add-to-a-waitgroup-from-multiple-goroutines

