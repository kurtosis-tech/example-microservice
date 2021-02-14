## 1.0.7
* Add validation to the API service's config file

## 1.0.6
* Make `DatastoreClient` time out requests after 2 seconds (low timeout, so tests that use timeouts run quickly)

## 1.0.5
* Use the `go` Docker image only for building, and use Alpine for execution; leads to a size reduction of 396MB -> 13MB!

## 1.0.4
* Fix silent bugs that I didn't notice before I added CI!

## 1.0.3
* Add `go test` to build.sh, to force building of Go files

## 1.0.2
* Bugfix in buildscript
* Add `.dockerignore` file

## 1.0.1
* Add CircleCI
* Refactored buildscript to do both `build` and `publish`
* Add `.dockerignore`, and a check in the buildscript to ensure it exists

## 1.0.0
* Init commit
* Add API container image
