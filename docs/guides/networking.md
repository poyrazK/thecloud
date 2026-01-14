# Networking Guide

This guide covers networking and port mapping features in The Cloud.

## Port Mapping

Port mapping allows you to expose container ports to your local machine, similar to the `-p` flag in Docker.

### How to use
When launching an instance, use the `-p` or `--port` flag with the format `hostPort:containerPort`.

```bash
cloud compute launch --name web --image nginx:alpine --port 8080:80
```

### Accessing your service
Once launched, if the status is `RUNNING`, you can access the service via `localhost`.

You can see the access URLs by listing your instances:
```bash
cloud compute list
```
Output:
```
┌──────────┬──────────┬──────────────┬─────────┬────────────────────┐
│    ID    │   NAME   │    IMAGE     │ STATUS  │       ACCESS       │
├──────────┼──────────┼──────────────┼─────────┼────────────────────┤
│ a1b2c3d4 │ web      │ nginx:alpine │ RUNNING │ localhost:8080->80 │
└──────────┴──────────┴──────────────┴─────────┴────────────────────┘
```

### Multiple Ports
You can map multiple ports by separating them with a comma (max 10):
```bash
cloud compute launch --name dual --image my-app --port 8080:80,3000:3000
```

## Virtual Private Cloud (VPC)
VPCs provide logically isolated sections of The Cloud network. All instances must belong to a VPC.

### Create a VPC
```bash
cloud vpc create --name prod-vpc
```

## Subnets
Subnets allow you to partition your VPC into smaller segments.

### Create a Subnet
```bash
cloud vpc create-subnet --vpc-id <vpc-id> --name private-1 --cidr 10.0.1.0/24
```

## Security Groups
Security Groups act as virtual firewalls for your instances, controlling inbound and outbound traffic.

### Create a Group
```bash
cloud sg create --vpc-id <vpc-id> web-sg --description "Web Traffic"
```

### Add Rules
Rules allow specific traffic based on protocol, port, and CIDR.
```bash
cloud sg add-rule <sg-id> --direction ingress --protocol tcp --port-min 80 --port-max 80 --cidr 0.0.0.0/0
```

### Manage Associations
Apply groups to instances to protect them.
```bash
cloud sg attach <instance-id> <sg-id>
cloud sg detach <instance-id> <sg-id>
```
