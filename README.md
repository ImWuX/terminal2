# Terminal2
Terminal2 is a modern web based terminal.

**NOTE** Terminal2 has only been tested on Arch Linux. It is designed to work on most linux distros. It will not work on Windows or MacOS.

## Build
1) Git clone the terminal repository. `git clone`
2) Build the terminal server. `go build`
3) Install frontend dependencies. `npm install`

## Config
Once Terminal2 is built, it is ready to go. You can simply run the terminal2 executable and navigate to `127.0.0.1:4100`. If you would like to install terminal as a daemon check out the [terminal2.service](./terminal2.service) file.

Terminal2 provides command-line options for configuration:  
`--database <path>` sets the location of the SQLite3 database.  
`--bind <address>` sets the bind address for the webserver. Ex. 127.0.0.1:3000  
`--webdir <path>` sets the location of the static web files.  
`--shell <shell>` sets the shell which terminal2 will use.  
**NOTE** the command-line options are also exposed as environment variables with the same name prefixed with `TERMINAL2_` and all uppercase. (`--var` would be `TERMINAL2_VAR`).

## Usage
The terminal2 executable doubles as a CLI for terminal2 sessions. The executable is automatically exposed in the PATH of any terminal2 session. CLI options:  
`--download <file>` downloads a file through the browser.  
`--theme <id>` changes the theme.  

### Themes
Themes are currently only configurable through manually editing the SQLite3 database. Themes are changed using the CLI, described [here](#usage).

### File Transfer
Upload works by simply dropping a file into the terminal window. This will upload the file to the current directory. Note that upload will never overwrite a file, if it already exists in the directory it will append a modifier to the filename.  
Download works by running a CLI command, described [here](#usage).