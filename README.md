# Will it blend?

Will It Blend is a tool that wraps an arbitrary CLI command. It then creates a Twitch Prediction that will automatically resolve based on the result of your wrapped command. This could be something like your test for a library. It could be for a "First Try" check of a CLI tool you made to see if you nailed it on the first try without running it yet.

This tool is inspired by my own [Twitch channel](https://www.twitch.tv/cmgriffing) as well as [ThePrimeagen's](https://www.twitch.tv/theprimeagen) "First Try" predictions.

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

If you would like to manually install the binary, you can find it in the [Releases](https://github.com/cmgriffing/will-it-blend/releases) section of this repo.


## Usage

To run the tool, the only required argument is the command that you would like to wrap. You should wrap it in quotes if there are spaces in the command.

```
will-it-blend "npm run test"
```

## Flags

You can also configure various parts of the process using these optional flags.

- `--title` || `-t`: the title of the prediction (max 45 chars). Default: "Will it blend?"
- `--duration` || `-d`: amount of seconds to run the prediction for (30s -> 18000s/30m). Default: `30`
- `--success` || `-s`: string for success option (max 25 chars). Default: "Yes"
- `--failure` || `-f`: string for failure option (max 25 chars) Default: "No"
- `--token` || `-t`: __NOT RECOMMENDED__ Your Twitch API token. You can pass this flag if you want to avoid the OAuth flow. This flag is not recommended to be set live on screen, but if you want to store it in a config file for future use.
- `--port` || `-p`: The port for the local server for Twitch authentication
  Must be one of `3000`, `4242`, `6969`, `8000`, `8008`, `8080`, or `42069`. Default: `3000`
- `--config` || `-c`: path to config file for persistent configuration of flags: Default: `~/.config/.will-it-blend.yaml`

## Caveats

There will be some caveats to wrapping commands. For now, there is just one regarding colorized output.

### Colorized Output

Some commands detect whether they are running in a proper terminal and disable colorized output when not done so.

Various commands often have arguments to cause them to force colors in the output.

Examples:

- BSD `ls -a` would need the `-G` flag added.
  eg: `will-it-blend "ls -a -G"`

- GNU `ls -a` would need the `--colors=always` flag added.
  eg: `will-it-blend "ls -a --colors=always"`

Be sure to consule the manual or help file for the command you are running.

## Sponsorship

If you found this tool useful, I would greatly appreciate a sponsorship on GitHub to fuel even more fun ideas like this.

My sponsorship page can be found here: [https://github.com/sponsors/cmgriffing](https://github.com/sponsors/cmgriffing)

## Users

Are you a Twitch streamer who has used this tool? You are more than welcome to have your channel listed in this Users section. Please open an Issue titled "New User {{your channel name}}".

- [cmgriffing](https://www.twitch.tv/cmgriffing)
- maybe you?
