# KATOK

## Overview
**KATOK (Kafka Topic Kreator)** - is command-line admin tool for create and update  topics parameters  for Apache Kafka.  
As a source of parameters can be used  yaml config file or Hashicorp Consul.


## Build

Prerequisites:

- Go Compiler
- GNU Make

By default we produce a binary with all the supported drivers with the following command:

```shell
make build
```

## Usage

Get katok, either as a [packaged release](https://github.com/grizzlybite/katok/releases/latest),
as a [Docker image](https://hub.docker.com/r/grizzlybite/katok).

Use the `--help` flag to get help information.

```shell
❯ ./katok --help
Usage: katok [flags]

Flags:
  -h, --help                                   Show context-sensitive help.
      --consul-enabled="false"                 Use consul: true || false ($CONSUL_ENABLED).
      --consul-url="http://127.0.0.1:8500"     Set consul url ($CONSUL_URL).
      --consul-token="you-consul-acl-token"    Set consul acl token ($CONSUL_TOKEN).
      --consul-config-path="kafka/topics"      Set consul config path ($CONSUL_CONFIG_PATH).
      --config-path="./topics.yaml"            Set path to yaml config file ($CONFIG_PATH).
      --version                                Print version 
```

### Demo
Let's set up a local stand for experiments, launch Kafka, Zookeeper and Consul. 


```shell
git clone git@github.com:grizzlybite/katok.git
cd ./example
docker compose -f docker-compose.yaml up -d 
```
Service Kafka contains setting `KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://kafka:29092` which specify the correct hostname or IP address that clients can use to connect to Kafka.

**Add an entry to your /etc/hosts:**
```shell
127.0.0.1  localhost kafka
```

**Generate Consul ACL token (OPTIONAL):**  
Get SecretID.
```shell
docker exec -ti 74e92c8daf19 consul acl bootstrap

AccessorID:       576199c1-0b61-23b3-9c8a-28ce6a1ae434
SecretID:         1b2a6076-7c8c-ef3d-9171-87478f7a8f6c
Description:      Bootstrap Token (Global Management)
Local:            false
Create Time:      2024-12-10 21:44:19.216897993 +0000 UTC
Policies:
   00000000-0000-0000-0000-000000000001 - global-management
```


#### File provider
In the root of the repository, there is a topics.yaml default file that contains parameters for the topics that need to be created.
```yaml
kafka_brokers: ["192.168.200.249:9092"]
topics:
  - name: "amon_amarth_topic"
    retention.ms: "86400077"
    num.partitions: 12
  - name: "decapitated_topic"
    retention.ms: 86400078
    num.partitions: 6
  - name: "if_flames_topic"
    retention.ms: 17211130
  - name: "lamb_of_god_topic"
    retention.ms: 86400078
  - name: "insomnium_topic"
    retention.ms: 86400078
    num.partitions: 18
```
Run katok:
```shell
❯ ./katok
{"time":"2024-12-11T00:50:32.778183873+03:00","level":"INFO","msg":"Using file config provider"}
{"time":"2024-12-11T00:50:34.061323598+03:00","level":"INFO","msg":"Topic 'amon_amarth_topic' was successfully created."}
{"time":"2024-12-11T00:50:34.703786831+03:00","level":"INFO","msg":"Topic 'decapitated_topic' was successfully created."}
{"time":"2024-12-11T00:50:35.902599025+03:00","level":"INFO","msg":"Topic 'if_flames_topic' was successfully created."}
{"time":"2024-12-11T00:50:37.200198239+03:00","level":"INFO","msg":"Topic 'lamb_of_god_topic' was successfully created."}
{"time":"2024-12-11T00:50:39.005620775+03:00","level":"INFO","msg":"Topic 'insomnium_topic' was successfully created."}
```

If we run the program again,we will see that the topics have been updated:
```shell
❯ ./katok
{"time":"2024-12-11T00:56:23.197374619+03:00","level":"INFO","msg":"Using file config provider"}
{"time":"2024-12-11T00:56:23.271316294+03:00","level":"INFO","msg":"Successfully update parameters for 'amon_amarth_topic' topic."}
{"time":"2024-12-11T00:56:23.321178374+03:00","level":"INFO","msg":"Successfully update parameters for 'decapitated_topic' topic."}
{"time":"2024-12-11T00:56:23.395480224+03:00","level":"INFO","msg":"Successfully update parameters for 'if_flames_topic' topic."}
{"time":"2024-12-11T00:56:23.454560026+03:00","level":"INFO","msg":"Successfully update parameters for 'lamb_of_god_topic' topic."}
{"time":"2024-12-11T00:56:23.487910831+03:00","level":"INFO","msg":"Successfully update parameters for 'insomnium_topic' topic."}
```
#### Consul provider
To use consul as the configuration source,you need use cli flags or set the environment variables:
```shell
export CONSUL_ENABLED=true
export CONSUL_URL=http://127.0.0.1:8501
export CONSUL_CONFIG_PATH=kafka/config
export CONSUL_TOKEN=1b2a6076-7c8c-ef3d-9171-87478f7a8f6c
```

```shell
❯ ./katok
{"time":"2024-12-11T22:43:49.881437963+03:00","level":"INFO","msg":"Using Consul config provider"}
{"time":"2024-12-11T22:43:49.931620466+03:00","level":"INFO","msg":"Successfully update parameters for 'amon_amarth_topic' topic."}
{"time":"2024-12-11T22:43:49.937806081+03:00","level":"INFO","msg":"Successfully update parameters for 'decapitated_topic' topic."}
{"time":"2024-12-11T22:43:49.94341785+03:00","level":"INFO","msg":"Successfully update parameters for 'if_flames_topic' topic."}
{"time":"2024-12-11T22:43:49.948963194+03:00","level":"INFO","msg":"Successfully update parameters for 'lamb_of_god_topic' topic."}
{"time":"2024-12-11T22:43:49.955621615+03:00","level":"INFO","msg":"Successfully update parameters for 'insomnium_topic' topic."}
```
