# TODO

## v0.2.x

* ~~API versioning~~

* ~~entrypoints (load balancer)~~

* ~~show component port addresses~~

* ~~standard ports / port preservation~~

* ~~resource metadata, tags~~

* ~~meta timestamps~~

## v0.3.x

* ~~volume resizing~~

* ~~delete releases on component delete (bugfix)~~

* ~~special release changes: service updates, load balancer port removal~~

## v0.4.x

* custom deployments

* add flag to release to indicate whether restart is needed (if instance config has changed)

## v0.5.x

* show instance metrics

* internal ELBs

## v0.6.x

*will probably use this release to organize some docs / tests*

* use AWS meta endpoint to get required info

* show resource common on all API responses

* standardized errors, removing panics where possible

* model validations

* TTL on failed tasks

* TTL on old releases

* standardized logging

## v0.7.x

* HTTP Basic Auth on API

* self-signed SSL cert for API HTTPS

## v0.8.x

* cluster info API: node IPs, instance common, etc.  (eventually billing stuff)

## v0.9.x

* instance container logs / SSH

## v0.10.x

* simple component security

## v0.11.x

* SSL / DNS

## v0.12.x

* bluegreen release

<hr>

### Misc

* godep (dependency versioning)

* unit tests / godoc

* API best-practices / Swagger

* docs/ folder
