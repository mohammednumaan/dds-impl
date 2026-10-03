# dds-impl

implementations of patterns and concepts from dds by brendan burns

> these are for learning and is not meant to be used in production

## implementations

1. [`sidecar pattern`](./sidecar) 
   - config manager that polls a config source for changes, updates a local config file and signals the application to reload via `sighup`

1. [`ambassador pattern`](./ambassador/) 
   - ambassador proxy that splits client requests between stable prod and beta servers, forwarding each request to the selected upstream based on a configurable percentage.

1. [`replicated-lb`](./replicated-lb/)
   - replicated load balancer that spreads client requests across identical app replicas, forwarding each request to the next replica in round-robin order.

1. [`sharding`](./sharding/)
   - sharded key-value store that spreads keys across distinct shards, hashing each request key to pick the shard that owns it.
