# go-kamailio

A simple package for interacting with Kamailio's binrpc system.  Eventually, we
should expand this to support kamailio's other RPC mechanisms and perhaps some
generalized tooling.

## Other kamailio-related Go projects

  * [CyCoreSystems/asterisk-k8s-demo](https://github.com/CyCoreSystems/asterisk-k8s-demo):
    a complete demonstration system using Asterisk, Kamailio, ARI, and scaling
    on kubernetes
  * [CyCoreSystems/dispatchers](https://github.com/CyCoreSystems/dispatchers):
    an app which synchronizes kamailio's dispatcher list with a number of
    Kubernetes Service Endpoints.  It also includes an API service for verifying
    connections from those endpoints.
  * [CyCoreSystems/kamconfig](https://github.com/CyCoreSystems/kamconfig): a
    dynamic configuration templating system for running kamailio inside
    kubernetes.

