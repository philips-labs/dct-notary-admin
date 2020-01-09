# Docker Content Trust - Notary Admin

This API and webapp add the capability to manage your Docker Content Trust and notary certificates.

It allows you to create new **Target** certificates for your Docker repositories, as well authorizing delegates for the repository.

This way the certificates can be stored in a secured environment where backups are managed.

## API endpoints

| HTTP Method | URL                                               | description                                    |
| ----------- | ------------------------------------------------- | ---------------------------------------------- |
| GET         | [](https://localhost:8443/ping)                   | return pong                                    |
| GET         | [](https://localhost:8443/targets)                | retrieves all target keys                      |
| GET         | [](https://localhost:8443/targets/{id})           | retrieves a single target key                  |
| GET         | [](https://localhost:8443/targets/{id}/delegates) | retrieves all delegate keys for a given target |

## Prerequisites

For a easier development workflow it is recommended to install **CMake**.

| platform | install                  | url                                |
| -------- | ------------------------ | ---------------------------------- |
| Windows  | `choco install -y cmake` | [cmake-3.16.2-win64-x64.msi][]     |
| MacOSX   | `brew install cmake`     | [cmake-3.16.2-Darwin-x86_64.dmg][] |

### Accept Self signed certs in Google Chrome

For Google Chrome to accept the selfsigned certificates please enable the option `allow-insecure-localhost` by navigating to [](chrome://flags/#allow-insecure-localhost) in your address bar.

To only allow for the current certificate that is blocked type `thisisunsafe` with focus on the `Your connection is not private` page, the page will autorefresh once the full phrase is typed. In older versions of chrome you had to type `badidea` or `danger`.

## Build

```bash
make build
```

## Test

Run the tests.

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

[cmake-3.16.2-win64-x64.msi]: https://github.com/Kitware/CMake/releases/download/v3.16.2/cmake-3.16.2-win64-x64.msi "Download cmake-3.16.2-win64-x64.msi"
[cmake-3.16.2-darwin-x86_64.dmg]: https://github.com/Kitware/CMake/releases/download/v3.16.2/cmake-3.16.2-Darwin-x86_64.dmg "Download cmake-3.16.2-Darwin-x86_64.dmg"

## Run

Now you can start the server as following:

```bash
bin/dctna
```

> **NOTE:** you can pass the sandbox `notary-config.json` as following. `bin/dctna -notary-config-file ./notary-config.json`.

Or via the Make shorthand which also builds the solution.

```bash
make run
```

> **NOTE:** via make we will also use our sandboxed `notary-config.json` automatically to prevent you from messing arround with your current notary (Production) settings.
