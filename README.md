# Will it blend?

Will It Blend is a tool that wraps an arbitrary CLI command. It then creates a Twitch Prediction that will automatically resolve based on the result of your wrapped command. This could be something like your test for a library. It could be for a "First Try" check of a CLI tool you made to see if you nailed it on the first try without running it yet.

This tool is inspired by my own [Twitch channel](https://www.twitch.tv/cmgriffing) as well as [ThePrimeagen's](https://www.twitch.tv/theprimeagen) "First Try" predictions.

**Warning: Your channel must be affiliated or paternered to be able to have channel points that viewers can spend.**

## Installation

The releases are built by [goreleaser](https://github.com/goreleaser/goreleaser) installation script is built by [godownloader](https://github.com/goreleaser/godownloader)

To install, run this command. It will create a .bin folder where the script was run. You will need to add that to your PATH.

```
curl -o- -s https://raw.githubusercontent.com/cmgriffing/will-it-blend/main/install.sh | bash
```

You can also pass in a directory already on your PATH.

```
curl -o- -s https://raw.githubusercontent.com/cmgriffing/will-it-blend/main/install.sh | bash -s -- -b /usr/local/bin/
```

You could also use `go install`:
```
go install github.com/cmgriffing/will-it-blend
```

If you would like to manually install the binary, you can find it in the [Releases](https://github.com/cmgriffing/will-it-blend/releases) section of this repo.

## Usage

To run the tool, the only required argument is the command that you would like to wrap. You should wrap it in quotes if there are spaces in the command.

```
will-it-blend "npm run test"
```

## Flags

You can also configure various parts of the process using these optional flags.

- `--title` || `-t`: the title of the prediction (max 45 chars). Default: "Will it blend?"
- `--duration` || `-d`: amount of seconds to run the prediction for (30s -> 18000s/30m). Default: `60`
- `--success` || `-s`: string for success option (max 25 chars). Default: "Yes"
- `--failure` || `-f`: string for failure option (max 25 chars) Default: "No"
- `--token` || `-k`: __NOT RECOMMENDED__ Your Twitch API token. You can pass this flag if you want to avoid the OAuth flow. This flag is not recommended to be set live on screen, but if you want to store it in a config file for future use.
- `--port` || `-p`: The port for the local server for Twitch authentication.
  Must be one of `3000`, `4242`, `6969`, `8000`, `8008`, `8080`, or `42069`. Default: `3000`
- `--config` || `-c`: path to config file for persistent configuration of flags: Default: `~/.config/.will-it-blend.yaml`

## How It Works

1. You run the command (see Usage section) with a command of your own.
2. The tool spins up a local server for Twitch Oauth purposes.
3. You go through the auth process to grant permission to the tool to create and modify Predictions. This should only be required once.
4. Twitch redirects to a completion HTML page that then sends the token to this tool.
5. The tool creates a Prediction via the [Twitch Predictions API](https://dev.twitch.tv/docs/api/predictions/).
6. Your wrapped command runs.
7. When your wrapped command completes, the tool uses the exit code of your command to resolve the prediction accordingly. Any non-zero exit code is a failure, otherwise success.

## Caveats

There will be some caveats to wrapping commands. For now, there is just one regarding colorized output.

### Colorized Output

Some commands detect whether they are running in a proper terminal and disable colorized output when not. Various commands often have arguments to cause them to force colors in the output.

Examples:

- BSD `ls -a` would need the `-G` flag added.
  eg: `will-it-blend "ls -a -G"`

- GNU `ls -a` would need the `--colors=always` flag added.
  eg: `will-it-blend "ls -a --colors=always"`

Be sure to consult the manual or help file for the command you are running.

## Sponsorship

If you found this tool useful, I would greatly appreciate a sponsorship on GitHub to fuel even more fun ideas like this.

My sponsorship page can be found here: [https://github.com/sponsors/cmgriffing](https://github.com/sponsors/cmgriffing)

## Users

Are you a Twitch streamer who has used this tool? You are more than welcome to have your channel listed in this Users section. Please open an Issue titled "New User {{your channel name}}".

- [cmgriffing](https://www.twitch.tv/cmgriffing)
- maybe you?

## License

Copyright 2023 Chris Griffing

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the “Software”), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
