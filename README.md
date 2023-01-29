# sphidewin

Hide/Unmap windows in X11.

## Features

- Hides any newly spawned window with the given property.
- Can also hide previously spawned windows.
- Automatically unhides/maps them on exit.

## Installation

Assuming [proper golang setup](https://go.dev/doc/install), simply clone this
repo, change directory to repo's directory and run the following command:

```console
$ go install
```

## Usage

```console
$ sphidewin --help
sphidewin [options] WM_CLASS

where [options] are:
  -h    Show help message
  -help
        Show help message
  -p    Unmap previously spawned windows
```

## Example

Running the above command

```console
$ sphidewin chromium
```

And opening `chromium` shows the following output:

```console
$ sphidewin chromium
2023/01/29 22:48:59 [chromium Chromium] 48234498 unmapped
```

And the chromium window is never shown.

Press `Ctrl-C` to unhide the chromium window.

```console
$ sphidewin chromium
2023/01/29 22:48:59 [chromium Chromium] 48234498 unmapped
^C2023/01/29 22:55:19 Mapped 48234498
```

## License

GPLv3. See [COPYING](COPYING).
