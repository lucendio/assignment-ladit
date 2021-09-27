SRE take-home test - LADIT
==========================


## Code

* Golang v1.17
* located in `./src`

### Reasoning

* it has been some time since I wrote Go, so I welcome the occasion
* for the purpose of simplicity, state will only reside in-memory, neither a backing service to manage
  the state outside of the service nor an interface to such service is being implemented
* due to some time constraints, it was decided to pull in dependencies for all the heavy lifting
  * API server
  * configuration
  * testing
* CIDR is used as identifier (not for instance the combination of CIDR and TTL), which means posting
  the same CIDR will have no effect
* at first, I wanted to decouple blocklist and store to ease adaption for a backing service to manage
  state externally, but then I realized that it would mean to write tests for an additional abstraction
  layer which is not needed for now
* although being required, it was decided to not include the IP blocking middleware in the `/healthcheck` 
  endpoint, because if a certain CIDR would be set to block, it might render the service unheathy from
  outer context's perspective that is tasked to keep the service up and running, e.g. a kubelet.
  Alternatively, one could just exclude private IPs (`IP.IsPrivate()`) or introduce a (configurable)
  list of unblockable IP ranges.
* adding a new CIDR on `/block` results in *201 Created* instead of *200 OK* as requested
* additional time constraints resulted in tests only to verify the HTTP API behaviour, instead of
  integration tests (e.g. for `main.go`) and unit tests (e.g. for `blocking/blocklist.go`)


## Containerfile

* three stages
* only built and tested with Docker so far


## Infrastructure

* used recommended Terraform provider: https://registry.terraform.io/providers/kreuzwerker/docker
* located in `./terraform`
* no Terraform state management whatsoever, since everything is (currently) applied locally
