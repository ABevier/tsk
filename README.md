# go-tasks

TODO:
 - Is Task the Right word?
 - Check OK value when reading channels
 - Should we have a logger?
 - readme
 Batching:
 - configuration
     - validate config and set defaults
 - comments
 - examples: SQS, SQL
 ---------
 Worker Queue (control max concurrency):
 - handle context cancellation better (pass it through all the way)
   - add worker id to the context
 - configuration
     - validate config and set defaults
