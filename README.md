# tsk

TODO:
 - readme
 - comments / documentation
 Batching:
 - configuration
     - validate config and set defaults
 - close batch
 - fix blocking when context is cancelled
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

