# DC-Plugged

A CLI tool for making it easier to work with plugins in site repos.

## Installation

- Here's how to [install go](https://go.dev/doc/install).
- Add the following lines to your `.zshrc`, `.bashrc` or `.profile`:
	- `export GOPATH=$HOME/go`
	- `export PATH=$PATH:$GOROOT/bin:$GOPATH/bin`

- `$ git clone https://github.com/designcontainer/dc-plugged`
- `$ cd dc-plugged`
- `$ go install`

## Nice to know

The script assumes you have your plugins located at `~/plugins/`.

The script uses you current directory name to find which plugin folder to apply changes to.

If your CWD is `~/sites/financepeople/wp-content/plugins/dc-post-grid/`, dc-plugged will make changes to `~/plugins/dc-post-grid/`.

## Usage

- Run `$ dc-plugged` to see all commands
- Create a new branch in plugin repo
	- `$ dc-plugged nb feat/cool-new-feature`
- Stage changes
	- `$ dc-plugged sc`
- Update version numbers in multiple files
	- `$ dc-plugged uv --files package-lock.json,dc-post-grid.php patch`
- Update version numbers in package.json only
	- `$ dc-plugged uv patch`
