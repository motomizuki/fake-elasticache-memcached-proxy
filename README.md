# fake-elasticache-memcached-proxy
A simple memcached proxy to emulate AWS Elasticache configuration endpoint.

## Usage

```
docker-compose up -d
```

Environments
```
FAKE_HOST: 
    type: string
    description: Listen host
    default: 0.0.0.0
FAKE_PORT: 
    type: string
    description: Listen port
    default: 11211
FAKE_PROXY_MEMCACHED_HOST:
    type: string
    description: Memcached host to which to proxy requests
    required: true
FAKE_PROXY_MEMCACHED_PORT:
    type: string
    description: Memcached port to which to proxy requests
    required: true
FAKE_CLUSTER_NODES: 
    type: string
    description: List of memcached nodes returned from configuration endpoint. format is ${host}|${ip}|${port} joined by space
    example:
        single node: "localhost|127.0.0.1|11211"
        multiple node: "localhost|127.0.0.1|11211 localhost|127.0.0.1|11211"
FAKE_TRACE: 
    type: bool
    description: Flag to control the output of the trace log
    default: false
```

## configuration endpoint

ref: https://docs.aws.amazon.com/AmazonElastiCache/latest/mem-ug/AutoDiscovery.HowAutoDiscoveryWorks.html

1. client that using Auto Discovery request `config get cluster` command to configuration endpoint
2. elasticache node return node list of cluster
3. if the node does not support `config get cluster` command, client try to get a key `AmazonElastiCache:cluster`.
4. Both 2. and 3. fails, client give up to connect memcached cluster.

`fake-elasticache-memcached-proxy` emulate 2.  
But you can also emulate configuration endpoint behavior to set `AmazonElastiCache:cluster`

VALUE format is 
```
${version}\n
${nodes}\n
```
current `${version}` is `5` and `${nodes}` format is same as `FAKE_CLUSTER_NODES`