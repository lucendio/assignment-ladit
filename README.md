SRE take-home test - LADIT
==========================

## Code

* Golang v1.17

### Reasoning

* it has been some time since I wrote Go, so I welcome the occasion
* for the purpose of simplicity, state will only reside in-memory, neither a backing service to manage
  the state outside of the service nor an interface to such service is being implemented
* due to some time constraints, it was decided to pull in dependencies for all the heavy lifting
  * API server
  * configuration
* CIDR is used as identifier (not for instance the combination of CIDR and TTL), which means posting
  the same CIDR will have no effect
* at first, I wanted to decouple blocklist and store to ease adaption for a backing service to manage
  state externally, but then I realized that it would mean to write tests for an additional abstraction
  layer which is not needed for now


## Containerfile


## Infrastructure

* Terraform provider: https://registry.terraform.io/providers/kreuzwerker/docker