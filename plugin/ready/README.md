# ready

## Name

*ready* - enables a readiness check HTTP endpoint.

## Description

By enabling *ready* an HTTP endpoint on port 8181 will return 200 OK, when all plugins that are able
to signal readiness have done so. If some are not ready yet the endpoint will return a 503 with the
body containing the list of plugins that are not ready.

Once a plugin has signaled it is ready it will not be queried again.

## Syntax

~~~
ready [ADDRESS]
~~~

*ready* optionally takes an address; the default is `:8181`. The path is fixed to `/ready`. The
readiness endpoint returns a 200 response code and the word "OK" when this server is ready. It
returns a 503 otherwise.

## Plugins

Any plugin wanting to signal readiness will need to implement the `ready.Readiness` interface by
implementing a method `Ready() bool` that returns true when the plugin is ready and false otherwise.

## Examples

Run another ready endpoint on <http://localhost:8091/ready>.

~~~ txt
. {
    ready localhost:8091
}
~~~
