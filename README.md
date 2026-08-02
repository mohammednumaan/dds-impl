# dds-impl

implementations of patterns and concepts from dds by brendan burns

> these are for learning and is not meant to be used in production

## implementations

1. [`sidecar pattern`](./sidecar) *(sidecar)*
   - config manager that polls a config source for changes, updates a local config file and signals the application to reload via `sighup`
