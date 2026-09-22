# pket

`pket` is a small package builder and package manager for applications and command-line tools. It creates compressed `.pkt` archives, verifies package contents with SHA-512, installs packages transactionally, and manages executable links.

## Requirements

- Go 1.27.1 or newer to build from source
- Linux, macOS, or Windows

## Build From Source

```sh
git clone https://github.com/badgardener/pket
cd pket
go build -o pket .
```

Run the executable directly or place it somewhere on your `PATH`:

```sh
./pket --version
```

## Usage

```text
pket [--verbose] <command> [arguments]
```

Commands:

```text
pket build <directory>       Build a package from a directory
pket install <package>       Install or update a package archive
pket list                    List installed packages
pket info <package>          Show package information
pket uninstall <package>     Uninstall a package by name or UID
pket --help                  Show help
pket --version               Show the version
```

Use `--verbose` before commands that support detailed callback output:

```sh
pket --verbose build ./my-app
pket --verbose install ./my-app-1.0.0.pkt
pket --verbose uninstall my-app
```

## Creating A Package

Create a directory containing the application files and a `pket-package.toml` manifest:

```text
my-app/
├── pket-package.toml
├── build/
│   └── my-app
├── LICENSE
└── README.md
```

Example manifest:

```toml
[package]
name = "My App"
pack = "my-app"
version = "1.0.0"
description = "An example packaged application."
authors = ["Example Author"]

[files]
base = "build"
assets = ["LICENSE", "README.md"]

[[executable]]
path = "my-app"
link = "my-app"

[install]
# preinstall = "@res/my-app --prepare"
# postinstall = "@res/my-app --version"
```

Build the package from the directory containing the manifest:

```sh
pket build ./my-app
```

The resulting archive is written inside the source directory using this name:

```text
<package.pack>-<package.version>.pkt
```

For the example above, the output is `my-app-1.0.0.pkt`.

### Manifest Fields

#### `[package]`

| Field         | Required | Description                                                            |
| ------------- | -------- | ---------------------------------------------------------------------- |
| `name`        | Yes      | Human-readable package name.                                           |
| `pack`        | Yes      | Package UID and archive name component. It must be one path component. |
| `version`     | Yes      | Three-part numeric version, such as `1.0.0`.                           |
| `description` | No       | Package description.                                                   |
| `authors`     | No       | Array of author names.                                                 |

#### `[files]`

| Field    | Required | Description                                                                      |
| -------- | -------- | -------------------------------------------------------------------------------- |
| `base`   | Yes      | Directory whose contents become the package `payload`.                           |
| `assets` | Yes      | Files or directories copied into `payload/pket-assets`. An empty array is valid. |

#### `[[executable]]`

At least one executable is required. Each entry contains:

| Field  | Required | Description                                                                  |
| ------ | -------- | ---------------------------------------------------------------------------- |
| `path` | Yes      | Path relative to the package `payload`.                                      |
| `link` | Yes      | Command name for the created executable link. It must be one path component. |

Executable paths may use `@res` in install commands and executable definitions. During installation, `@res` resolves to the package payload directory.

#### `[install]`

| Field         | Required | Description                                                       |
| ------------- | -------- | ----------------------------------------------------------------- |
| `preinstall`  | No       | Command offered for confirmation before the package is activated. |
| `postinstall` | No       | Command offered for confirmation after installation completes.    |

Install commands run through the platform shell with the installation directory as the working directory. They require two confirmations before execution.

## Installing A Package

```sh
pket install ./my-app-1.0.0.pkt
```

Installation performs these checks and actions:

1. The archive is extracted into a temporary directory.
2. Archive paths and links are checked to prevent extraction outside that directory.
3. Payload files are verified against the package SHA-512 manifest when present.
4. The package is prepared in a staging directory.
5. The staging directory is atomically activated as the installed package.
6. Executable links are created in the user executable directory.

If a package with the same UID is already installed, `pket` asks whether it should be updated. The previous installation is retained as a rollback source until activation succeeds.

Packages without SHA-512 sums require explicit confirmation before installation continues.

## Managing Packages

List installed packages:

```sh
pket list
```

Show package details by package name or UID:

```sh
pket info my-app
```

Remove a package:

```sh
pket uninstall my-app
```

Uninstalling also removes executable links recorded by the package. The package archive and installed files are removed after confirmation.

## Installed Files

Package data is stored under:

| Platform                     | Package directory                    |
| ---------------------------- | ------------------------------------ |
| Linux and other Unix systems | `~/.local/share/pket`                |
| macOS                        | `~/Library/Application Support/pket` |
| Windows                      | `~/AppData/Roaming/pket`             |

Each installed package contains:

```text
<package-uid>/
├── payload/
├── pket-manifest/
│   ├── pket-config.toml
│   ├── sha-512.sums
│   └── files.lst
└── <package-name>.pkt
```

On Linux, executable links are created in `~/.local/bin` when that directory is writable. macOS, Windows, and other platforms use the first suitable writable executable directory selected by `pket`.

## Package Format

A `.pkt` file is a gzip-compressed tar archive. It contains:

- `payload/`, containing application files
- `pket-manifest/pket-config.toml`, containing package metadata
- `pket-manifest/sha-512.sums`, containing the payload integrity hash

The builder uses parallel gzip compression, and the installer uses parallel gzip decompression where supported. Installation also verifies the payload before activation and avoids replacing the active installation until verification and staging complete.

## Troubleshooting

### `pket-package.toml` is missing

Add the manifest to the project root.

### An executable link cannot be created

Make sure the requested link name is not already used and that the selected executable directory is writable.

### A package update is declined

Run the install command again and confirm the update prompt. Existing installations are never replaced without confirmation.

### SHA-512 verification fails

Rebuild the package from the intended source directory. A modified archive or payload will not pass verification.

## License

See [LICENSE](LICENSE).
