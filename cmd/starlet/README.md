# Legacy Starlet command

For the dedicated command-line application, use [StarCLI](https://github.com/1set/starcli). This older command is a **separate Go module**; root-library APIs and test results do not establish its behavior.

Its checked-in [go.mod](go.mod) declares Go 1.18, Starlet `v0.1.0`, Starlight `v0.1.0`, and `go.starlark.net v0.0.0-20240123142251-f86470692795`. These differ from the current root library. Installing the command does not select the root checkout or make current `eval` helpers available.

## Installation and use

The retained installation command is:

```sh
go install github.com/1set/starlet/cmd/starlet@latest
```

`@latest` selects the command module's published version, which may differ from this source. Inspect the installed binary with `go version -m /path/to/starlet`. Installation and interactive execution have not been revalidated as part of this root-library documentation update.

Run a supplied expression as a script:

```sh
starlet -c 'print(1 + 2)'
```

Run a file, or start the REPL with no arguments:

```sh
starlet example.star
starlet
```

An offline REPL example:

```text
>>> 1 + 2
3
```

Use `starlet --help` for flags. `--include` selects the root for file-loaded modules (default `.`); `--preload` and `--lazyload` select builtin modules. The command defaults to all builtin preloads and loaders, including modules with host effects. Run only trusted scripts; it is not a sandbox. For embedding against the current root library, see the [Machine usage guide](../../USAGE.md).
