# Futures

TODO:
- NewFromFunction

## Why?

Using a channel as a future doesn't work when the value will be read by multiple consumers. A value can only
be read by a channel once before it is consumed and lost.

Futures are useful when multiple requestors ask for the same data simultaneously.  The future can be completed once
and all requestors will see the same value.
