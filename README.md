# tsk

TODO:
 - readme
 - repackage things
 Batching:
 - configuration
     - validate config and set defaults
 - close batch
 - fix blocking when context is cancelled
 - comments
 - examples: SQS, SQL
 ---------
 Worker Queue (control max concurrency):
 - handle context cancellation better (pass it through all the way)
   - add worker id to the context
 - configuration
     - validate config and set defaults
 ----------
 - Futures?
 - Retries?
 - Parallelize?

