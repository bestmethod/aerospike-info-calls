# aerospike-info-calls

## Make info calls to aerospike

### NOTE: Does not support TLS

### bin/aerospike-info-check-time <- linux compiled binary, just download, chmod and run

### src/main.go <- golang source code

## USAGE:
```
Usage: aerospike-info-check-time -h [HOSTNAME[:IP]] [-u USERNAME -p PASSWORD] [-n NODE_IP,NODE_IP,...] [-t TIMEOUT_MS] COMMAND

Example: Query and seed from only 'localhost':
	aerospike-info-check-time -n 127.0.0.1 'services'

Example: Query all nodes for 'services', connect to localhost:
	aerospike-info-check-time 'services'

Example: As above, timeout 300 seconds
	aerospike-info-check-time -t 300000 'services'

Example: Connect to 10.0.0.6:3000, use user/pass and get output only from 2 nodes in the cluster
	aerospike-info-check-time -h 10.0.0.6:3000 -u superman -p crypton -n 10.0.0.5,10.0.0.6 -t 300000 'services'
```

## OUTPUT
```
Connected in 0.007s

NODE BB9020011AC4202 (127.0.0.1:3000) RESPONDED IN 0.000s:
172.17.0.3:3000;172.17.0.4:3000

NODE BB9030011AC4202 (172.17.0.3:3000) RESPONDED IN 0.000s:
172.17.0.4:3000;172.17.0.2:3000

NODE BB9040011AC4202 (172.17.0.4:3000) RESPONDED IN 0.000s:
172.17.0.3:3000;172.17.0.2:3000
```
