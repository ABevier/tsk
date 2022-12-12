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
 Task Queue (control max concurrency):
 - Fix races
 - handle context cancellation better (pass it through all the way)
   - add worker id to the context
 - configuration
     - validate config and set defaults
 ----------
 - Retries?

