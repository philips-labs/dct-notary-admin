# Docker Content Trust - Notary Admin

This API and webapp add the capability to manage your Docker Content Trust and notary certificates.

It allows you to create new **Target** certificates for your Docker repositories, as well authorizing delegates for the repository.

This way the certificates can be stored in a secured environment where backups are managed.

## Prerequisites

For a easier development workflow it is recommended to install **CMake**.

| platform | install                  | url                                |
| -------- | ------------------------ | ---------------------------------- |
| Windows  | `choco install -y cmake` | [cmake-3.16.2-win64-x64.msi][]     |
| MacOSX   | `brew install cmake`     | [cmake-3.16.2-Darwin-x86_64.dmg][] |

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
