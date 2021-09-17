# network-monitor 

This is a monitor that monitors the network by calling a third-party API.

# Overview 

Some times our application will meet some error while running inside our customer's device, some times it is about network, so I have to verify the problem. But our users are not very professional in information technology, and they are very impatient. If users are required to provide more detailed information about the error, they will sometimes be very angry, so the best way is to figure out what happened by ourself. 

Since we are usually not in the city where the user is located, we need to have an agent in that city. Some companies provide such agents, such as 17ce. 

This program connects to third-party APIs and performs continuous monitoring based on configuration files, and exports the monitoring data to a format that Prometheus can recognize, so that Grafana can conveniently use these data to draw pictures.

# Build

```bash
$ git clone https://github.com/ChaunceyShannon/network-monitor
$ cd network-monitor
$ go build .
```