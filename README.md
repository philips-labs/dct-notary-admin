# Docker Content Trust - Notary Admin

[![Continuous Integration](https://github.com/philips-labs/dct-notary-admin/workflows/Continuous%20Integration/badge.svg)](https://github.com/philips-labs/dct-notary-admin/actions?query=workflow%3A"Continuous+Integration"+branch%3Adevelop)
[![codecov](https://codecov.io/gh/philips-labs/dct-notary-admin/branch/develop/graph/badge.svg)](https://codecov.io/gh/philips-labs/dct-notary-admin)

This API and webapp add the capability to manage your Docker Content Trust and notary certificates.

It allows you to create new **Target** certificates for your Docker repositories, as well authorizing delegates for the repository.

This way the certificates can be stored in a secured environment where backups are managed.

This project makes use of a [Notary sandbox][notaryforksandbox] which is an in progress development setup, which is intended to be contributed [upstream][notary]. The [Fork][notaryfork] is by no means a hard Fork of [Notary][notary] and is solely there to bridge a period of time to get this back in the upstream.

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

For a easier development workflow it is recommended to install **CMake**. To interact with [Hashicorp Vault] the `vault cli` is convenient.

| platform | install                  | url                              |
| -------- | ------------------------ | -------------------------------- |
| Windows  | `choco install -y cmake` | [cmake-3.16.2-win64-x64.msi]     |
| MacOSX   | `brew install cmake`     | [cmake-3.16.2-Darwin-x86_64.dmg] |
| Windows  | `choco install -y vault` | [vault_1.4.2_windows_amd64.zip]  |
| MacOSX   | `brew install vault`     | [vault_1.4.2_darwin_amd64.zip]   |

### Accept Self signed certs in Google Chrome

For Google Chrome to accept the selfsigned certificates please enable the option `allow-insecure-localhost` by navigating to [](chrome://flags/#allow-insecure-localhost) in your address bar.

To only allow for the current certificate that is blocked type `thisisunsafe` with focus on the `Your connection is not private` page, the page will autorefresh once the full phrase is typed. In older versions of chrome you had to type `badidea` or `danger`.

## Run the sandbox

To run in an isolated environment to do some testing you should run the sandbox. The sandbox connects to a notary server and registry in the docker-compose setup.

Initializing the notary git submodule.

```bash
git submodule update --init --recursive
```

Build the sandbox

```bash
make build-sandbox
```

Run the sandbox

```bash
make run-sandbox
```

To provision the notary sandbox with some signed images you can use the `bootstrap-sandbox` make target (optional).

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

## Run vault development server

To boot the [Hashicorp Vault](https://www.vaultproject.io/) development server run the following. Requires vault installed, (e.g. `brew install vault`).

```bash
docker-compose -f vault/docker-compose.dev.yml up -d
vault/prepare.sh dev
```

`prepare.sh` boots vault server and provisions the secret engine with required policies, secret engines etc.

The vault admin dashboard is available at [http://localhost:8200].

The token can be found in the server logs.

```bash
docker-compose -f vault/docker-compose.dev.yml logs | grep "Root Token"
```

## Build binary (api)

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

> The API utilizes Hashicorp vault to generate and store passwords for private keys. The endpoint to Hashicorp vault can be configured via the environment variable `VAULT_ADDR` or as a commandline flag. The default value points to [http://localhost:8200] (the address of the development server).

### API Server

Now you can start the API server as following:

```bash
# environment variable
export VAULT_ADDR=http://localhost:8200
bin/dctna --config .notary/config.json
```

Alternatively you provide the vault server address as parameter.

```bash
# commandline option
bin/dctna --vault-addr http://localhost:8200 --config .notary/config.json
```

> **NOTE:** you can pass the sandbox `.notary/config.json` as above, without this setting the default notary folder will be used (`$USER/.natary/config.json`).

Or via the Make shorthand which also builds the solution, which will use the sandbox config for notary.

```bash
make run
```

> **NOTE:** via make we will also use our sandboxed `.notary/config.json` automatically to prevent you from messing arround with your current notary (Production) settings.

### Web Frontend

```bash
cd web
yarn install && yarn start
```

## Testing end to end

Now you can create new targets for signing docker images [http://localhost:3000] using the webinterface.

E.g.:

- Target: `localhost:5000/nginx`
- Target: `localhost:5000/stuff`

Then on one of the targets we will authorize our personal delegation key. If you don't have one yet you can generate it via the docker trust cli.

```bash
docker trust key generate johndoe --dir ~/.docker/trust
```

Then simply copy the contents of the public key to your clipboard.

```bash
cat ~/.docker/trust/johndoe.pub | pbcopy
```

In the webinterface you can now add your delegation on the target `localhost:5000/nginx`.

| name     | key                      |
| -------- | ------------------------ |
| john_doe | paste_clipboard_contents |

Now to be able to sign an image all signing keys have to be available on your local system. In Notary v2 this will be improved to also be able to work with remote signing keys. You will only need the passphrase for your delegation key.

This will now allow us to sign docker images for `localhost:5000/nginx`. In below example we first pull an image from the public registry. Then tag it to push to our sandbox registry. Then we enable content trust and configure our sandbox notary endpoint. Then we use the `dctna` cli to download the signing keys and tuf metadata. Upon pushing to the repository you will be prompted for the password of your signing key.

```bash
docker pull nginx:alpine
docker tag nginx:alpine localhost:5000/nginx:alpine
export DOCKER_CONTENT_TRUST=1 DOCKER_CONTENT_TRUST_SERVER=https://localhost:4443
bin/dctna --server-address https://localhost:8443 localhost:5000/nginx
docker push localhost:5000/nginx:alpine
```

[cmake-3.16.2-win64-x64.msi]: https://github.com/Kitware/CMake/releases/download/v3.16.2/cmake-3.16.2-win64-x64.msi "Download cmake-3.16.2-win64-x64.msi"
[cmake-3.16.2-darwin-x86_64.dmg]: https://github.com/Kitware/CMake/releases/download/v3.16.2/cmake-3.16.2-Darwin-x86_64.dmg "Download cmake-3.16.2-Darwin-x86_64.dmg"
[Hashicorp Vault]: https://vaultproject.io "Manage secrets and protect sensitive data"
[vault_1.4.2_windows_amd64.zip]: https://releases.hashicorp.com/vault/1.4.2/vault_1.4.2_windows_amd64.zip "Download vault_1.4.2_windows_amd64.zip"
[vault_1.4.2_darwin_amd64.zip]: https://releases.hashicorp.com/vault/1.4.2/vault_1.4.2_darwin_amd64.zip "Download vault_1.4.2_darwin_amd64.zip"
[notary]: https://github.com/theupdateframework/notary "Notary is a project that allows anyone to have trust over arbitrary collections of data"
[notaryfork]: https://github.com/philips-labs/notary/blob/feature/sandbox "This Fork is only to support the submodule which contains the sandbox setup"
[notaryforksandbox]: https://github.com/philips-labs/notary/blob/feature/sandbox/docker-compose.sandbox.yml "Notary docker-compose sandbox setup"
[http://localhost:8200]: http://localhost:8200 "Vault address"
[http://localhost:3000]: http://localhost:3000 "DCTNA dashboard address"
