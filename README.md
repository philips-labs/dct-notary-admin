# Docker Content Trust - Notary Admin

[![CI](https://github.com/philips-labs/dct-notary-admin/workflows/CI/badge.svg)](https://github.com/philips-labs/dct-notary-admin/actions?query=branch%3Adevelop)
[![codecov](https://codecov.io/gh/philips-labs/dct-notary-admin/branch/develop/graph/badge.svg)](https://codecov.io/gh/philips-labs/dct-notary-admin)

This API and webapp add the capability to manage your Docker Content Trust and notary certificates.

It allows you to create new **Target** certificates for your Docker repositories, as well authorizing delegates for the repository.

This way the certificates can be stored in a secured environment where backups are managed.

This project makes use of a [Notary sandbox][NotaryForkSandbox] which is an in progress development setup, which is intended to be contributed [upstream][Notary]. The [Fork][NotaryFork] is by no means a hard Fork of [Notary][Notary] and is solely there to bridge a period of time to get this back in the upstream.

## API endpoints

| HTTP Method | URL                                                                                                                          | description                                    |
| ----------- | ---------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------- |
| GET         | [https://localhost:8443/ping](https://localhost:8443/ping)                                                                   | return pong                                    |
| GET         | [https://localhost:8443/targets](https://localhost:8443/targets)                                                             | retrieves all target keys                      |
| POST        | [https://localhost:8443/targets](https://localhost:8443/targets)                                                             | creates a new target and keys                  |
| GET         | [https://localhost:8443/targets/{id}](https://localhost:8443/targets/{id})                                                   | retrieves a single target key                  |
| GET         | [https://localhost:8443/targets/{id}/delegations](https://localhost:8443/targets/{id}/delegations)                           | retrieves all delegate keys for a given target |
| POST        | [https://localhost:8443/targets/{id}/delegations](https://localhost:8443/targets/{id}/delegations)                           | add a new delegation to the given target       |
| DELETE      | [https://localhost:8443/targets/{id}/delegations/{delegation}](https://localhost:8443/targets/{id}/delegations/{delegation}) | remove a delegation from the given target      |

## Prerequisites

For a easier development workflow it is recommended to install **CMake**.

| platform | install                  | url                                |
| -------- | ------------------------ | ---------------------------------- |
| Windows  | `choco install -y cmake` | [cmake-3.16.2-win64-x64.msi][]     |
| MacOSX   | `brew install cmake`     | [cmake-3.16.2-Darwin-x86_64.dmg][] |

### Accept Self signed certs in Google Chrome

For Google Chrome to accept the selfsigned certificates please enable the option `allow-insecure-localhost` by navigating to [](chrome://flags/#allow-insecure-localhost) in your address bar.

To only allow for the current certificate that is blocked type `thisisunsafe` with focus on the `Your connection is not private` page, the page will autorefresh once the full phrase is typed. In older versions of chrome you had to type `badidea` or `danger`.

## Run the sandbox

To run in an isolated environment to do some testing you should run the sandbox. The sandbox connects to a notary server and registry in the docker-compose setup.

Initializing the notary git submodule.

```bash
git submodule init
git submodule update
```

Build the sandbox

```bash
make build-sandbox
```

Run the sandbox

```bash
make run-sandbox
```

To provision the notary sandbox with some signed images you can use the `bootstrap-sandbox` make target.

```bash
make bootstrap-sandbox
```

To play with the notary and docker trust cli you can open the shell for the sandbox. [Signing docker images using docker content trust](https://marcofranssen.nl/signing-docker-images-using-docker-content-trust/)

```bash
docker-compose -f notary/docker-compose.sandbox.yml -f docker-compose.yml exec sandbox sh
```

To shutdown the sandbox you can run the `stop-sandbox` make target.

```bash
make stop-sandbox
```

## Build binary

```bash
make build
```

## Test

To run the tests, make sure to run `make stop-sandbox` first (tests are also reusing the same sandbox which require a clean env).

```bash
make test
```

Run the tests with coverage.

```bash
make coverage-out
```

Check the coverage report in your browser.

```bash
make coverage-html
```

## Run

> For the API to provide the key credentials following environment variables have to be set. Later on different credentials for different keys will be dynamically loaded from a secure storage.

```bash
export NOTARY_ROOT_PASSPHRASE=test1234
export NOTARY_TARGETS_PASSPHRASE=test1234
export NOTARY_SNAPSHOT_PASSPHRASE=test1234
```

Now you can start the server as following:

```bash
bin/dctna
```

> **NOTE:** you can pass the sandbox `.notary/config.json` as following. `bin/dctna --config .notary/config.json`.

Or via the Make shorthand which also builds the solution.

```bash
make run
```

> **NOTE:** via make we will also use our sandboxed `.notary/config.json` automatically to prevent you from messing arround with your current notary (Production) settings.

[cmake-3.16.2-win64-x64.msi]: https://github.com/Kitware/CMake/releases/download/v3.16.2/cmake-3.16.2-win64-x64.msi "Download cmake-3.16.2-win64-x64.msi"
[cmake-3.16.2-darwin-x86_64.dmg]: https://github.com/Kitware/CMake/releases/download/v3.16.2/cmake-3.16.2-Darwin-x86_64.dmg "Download cmake-3.16.2-Darwin-x86_64.dmg"
[Notary]: https://github.com/theupdateframework/notary "Notary is a project that allows anyone to have trust over arbitrary collections of data"
[NotaryFork]: https://github.com/philips-labs/notary/blob/feature/sandbox "This Fork is only to support the submodule which contains the sandbox setup"
[NotaryForkSandbox]: https://github.com/philips-labs/notary/blob/feature/sandbox/docker-compose.sandbox.yml "Notary docker-compose sandbox setup"
