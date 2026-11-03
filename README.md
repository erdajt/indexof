# indexof

a terminal browser for open directories. browse files, download them, or open them in your editor.

## Install

you'll need go 1.22 or newer and an editor on your `PATH`. the default is `nvim`.

```sh
git clone https://github.com/erdajt/indexof.git
cd indexof
go mod tidy
go build -o indexof ./cmd/indexof
./indexof https://www.debian.org/CD/mirroring/
```

or install it directly:

```sh
go install github.com/erdajt/indexof/cmd/indexof@latest
indexof https://www.debian.org/CD/mirroring/
```

for `go install`, add `GOBIN` to your `PATH`, or `$(go env GOPATH)/bin` if you haven't set it.

## Keys

| key | action |
| --- | --- |
| `j` / `k`, `↓` / `↑` | down / up |
| `10j`, `5k` | move by a count |
| `l`, `→`, `enter` | enter a directory or open a file |
| `h`, `←`, `backspace` | go back through directories you've visited |
| `G` / `$` | last item |
| `v`, move, `y` | select a range and copy its urls |
| `d` | download without opening |
| `p` | downloads |
| `s` | settings |
| `c` | switch between the configured download folder and the current directory |
| `K` | toggle delete-on-exit |
| `r` / `q` | reload / quit |

in downloads: `space` pauses or resumes, `x` cancels, `c` clears finished jobs, and `p` or `esc` goes back.

if you wanna keep your downloads, turn delete-on-exit off with `K`. switching the destination with `c` doesn't turn cleanup off.

## Config

press `s` for settings. changes save to `~/.config/indexof/config.yaml`.

the default editor is `nvim`. downloads go to the system's temporary directory under `indexof`, and delete-on-exit is enabled by default.

themes: `default`, `tokyo`, `catppuccin`, `nord`, `gruvbox`, `dracula`.

## License

mit
