# go-tasks

TODO:
 - Is Task the Right word?
 - Check OK value when reading channels
 - Should we have a logger?
 Batching:
 - configuration
     - validate config and set defaults
 - comments
 - readme
 - examples: SQS, SQL
 ---------
 Worker Queue (control max concurrency):
 - initial implementation
 - backpressure 2 modes:
   - block (default)
   - return error (option)
