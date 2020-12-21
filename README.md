# dohboy
dohboy is a little [DNS-over-HTTPS](https://tools.ietf.org/html/rfc8484) relay.

This is currently a bit of a work-in-progress and not especially well-tested, but it does seem to do the thing. If you're looking to self-host a DOH relay and somehow stumble upon this, I'd suggest doing a little research as there are more mature solutions out there. 

### Building/Running
```
# go build
#
# ./dohboy -config=/my/config/file.yml
```

### Configuration
Configuration is done through a yaml file, the semantics of which are defined in [config.go](server/config.go).

#### Notes:
- Providing a TLSCertPath and TLSKeyPath will configure the server for HTTPS. If the plan is to run dohboy behind a reverse proxy and do SSL offloading there, leaving them empty will cause it serve up everything over http
- In specifying custom upstreams, there is a `NameRegex` field. For an incoming request, dohboy will compare the DNS question name against each regex pattern in the order that the upstreams have been configured, and will use the first matching upstream to resolve the msg. That way you can shunt off queries for *.local for instance to one target and everything else to another.

### Future Work
- Caching responses is currently implemented with standard HTTP caching, which for my small number of clients should probably perform pretty well. But it could be valuable to maintain a second in-memory layer of cached responses too.
- It might be nice to add some utils to collect metrics about average rtts to exchange messages with each upstream, caching stats, etc.
- The token-based rate-limit whitelist is a nice idea, but doesn't appear to work as well with firefox as I hoped. Maybe there's a better approach there, but I want to avoid ip-based whitelisting.
