package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/urfave/cli/v2"
)

/**
 * Helper functions
 */

func check(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func getHomeDir() (homeDir string) {
	homeDir, err := os.UserHomeDir()
	check(err)

	return
}

// The dir where all plugin repos are located.
var pluginsDir string = path.Join(getHomeDir(), "plugins")

// Returns the current working directory.
func getCWD() (dir string) {
	dir, err := os.Getwd()
	check(err)

	return
}

// Returns the plugin repo path based on CWD.
func getThePluginDir() (dir string) {
	dir = path.Base(getCWD())
	dir = path.Join(pluginsDir, dir)

	return
}

// Cheks if `e` is in `s`
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

/**
 * Action functions
 */

// Creates a new branch in `getThePluginDir()` with `branchName` as the name.
// Also checks out that branch.
func newBranch(branchName string) {
	// Open the repo so we can make changes to it
	repo, err := git.PlainOpen(getThePluginDir())
	check(err)

	// Open the worktree
	workTree, err := repo.Worktree()
	check(err)

	// Check if repo is dirty
	workTreeStatus, err := workTree.Status()
	if !workTreeStatus.IsClean() {
		fmt.Printf("%s is dirty.\n", getThePluginDir())
		fmt.Println("Please commit or stash your changes.")
		return
	}

	// Create a branch and store the new branch
	// Bash equivalent: git branch test
	refName := plumbing.NewBranchReferenceName(branchName)
	headRef, err := repo.Head()
	check(err)
	ref := plumbing.NewHashReference(refName, headRef.Hash())
	err = repo.Storer.SetReference(ref)
	check(err)

	// Checkout the new branch
	err = workTree.Checkout(&git.CheckoutOptions{Branch: ref.Name()})
	check(err)

	if err == nil {
		fmt.Printf("Created branch %s in %s\n", branchName, getThePluginDir())
	}
}

// Copies over the changes from the plugin dir in site repo to the plugin repo
// First it deletes all files in the plugin repo.
// Then it copies the files over using `cp`.
func stageChanges() {
	// Heres a list of ignored files and dirs we don't want to mess with.
	ignoredFiles := []string{".git", ".gitignore", "node_modules", ".env"}

	// Check if the plugin dir exists
	if _, err := os.Stat(getThePluginDir()); os.IsNotExist(err) {
		fmt.Printf("%s does not exist.\n", getThePluginDir())
		fmt.Printf(
			"Run `$ git clone %s/%[2]s ~/plugins/%[2]s`\n",
			"https://github.com/designcontainer",
			path.Base(getCWD()),
		)
		return
	}

	// Delete all files in the plugin repo that are not in the ignored list
	files, err := os.ReadDir(getThePluginDir())
	for _, dirEntry := range files {
		if !contains(ignoredFiles, dirEntry.Name()) {
			os.RemoveAll(path.Join(getThePluginDir(), dirEntry.Name()))
		}
	}

	// Copy over files from CWD to the plugin directory
	files, err = os.ReadDir(getCWD())
	check(err)
	for _, dirEntry := range files {
		if !contains(ignoredFiles, dirEntry.Name()) {
			var cp *exec.Cmd

			// Create the `cp` command based on if the dirEntry is a dir or not
			if dirEntry.IsDir() {
				cp = exec.Command(
					"cp",
					"-R", // Recursive flag, so we can copy dirs as well
					dirEntry.Name(),
					path.Join(getThePluginDir(), dirEntry.Name()),
				)
			} else {
				cp = exec.Command(
					"cp",
					dirEntry.Name(),
					path.Join(getThePluginDir(), dirEntry.Name()),
				)
			}

			// Define outputs
			cp.Stdout = os.Stdout
			cp.Stderr = os.Stderr

			// Run the command
			cp.Run()
		}
	}

	fmt.Printf(
		"Staged changes. They are ready to be commited in %s\n",
		getThePluginDir(),
	)
}

// Updates the plugin version number based on input level
// `level` can be: major, minor or patch
// It only changes version number in package.json by default.
// `files` is a comma seperated string of other files to change version in.
func updateVersionNumbers(level string, files string) {
	// Read the package.json file
	packageJSONBytes, err := os.ReadFile("package.json")
	check(err)

	// Unmarshal/parse the json into a map
	var packageJSONMap map[string]interface{}
	json.Unmarshal(packageJSONBytes, &packageJSONMap)
	oldVersion := packageJSONMap["version"].(string)

	// Increment the version number according to the level
	newVersion := strings.Split(packageJSONMap["version"].(string), ".")
	var num int

	// Get the index of the version number to increment
	var table = map[string]int{
		"major": 0,
		"minor": 1,
		"patch": 2,
	}
	index := table[level]

	// Update the version number
	num, err = strconv.Atoi(newVersion[index])
	check(err)
	newVersion[index] = fmt.Sprint(num + 1)

	newVersionStr := strings.Join(newVersion, ".")

	fmt.Printf(
		"Updated version from %s to %s.\n",
		packageJSONMap["version"],
		strings.Join(newVersion, "."),
	)

	// NOTICE:
	// We str replace the version number instead of stringifying our parsed
	// json, and writing that to the file, to avoid messing with formatting.

	// Replace the version number in the package.json file
	fileString := strings.Replace(
		string(packageJSONBytes), // Replace target
		oldVersion,               // Search for previous version
		newVersionStr,            // Replace with new version
		1,                        // Only replace the first instance
	)

	// Write to both the site repo and the plugin repo
	os.WriteFile("package.json", []byte(fileString), 0644)
	os.WriteFile(
		path.Join(getThePluginDir(), "package.json"),
		[]byte(fileString),
		0644,
	)

	// Loop trough the extra files and replace the version.
	fileList := strings.Split(files, ",")
	for _, file := range fileList {
		// Read the file
		fileBytes, err := os.ReadFile(file)
		check(err)

		fileString := strings.Replace(
			string(fileBytes), // Replace target
			oldVersion,        // Search for previous version
			newVersionStr,     // Replace with new version
			1,                 // Only replace the first instance
		)

		// Write to both the site repo and the plugin repo
		os.WriteFile(path.Join(file), []byte(fileString), 0644)
		os.WriteFile(
			path.Join(getThePluginDir(), file),
			[]byte(fileString),
			0644,
		)
	}
}

func main() {
	// Define destination for --files flag in update-version command
	var files string

	app := &cli.App{
		Name:    "DC-Plugged",
		Usage:   "Automate boring tasks related to working with plugins",
		Version: "1.0.0",
		Commands: []*cli.Command{
			{
				Name:      "new-branch",
				Aliases:   []string{"nb"},
				Usage:     "Creates a new branch in the plugin repo",
				ArgsUsage: "[new branch name]",
				Action: func(c *cli.Context) (err error) {
					branchName := c.Args().First()
					// Check if branchName is valid
					if branchName == "" {
						fmt.Println("Please provide a branch name")
						return
					}

					newBranch(branchName)
					return
				},
			},
			{
				Name:    "stage-changes",
				Aliases: []string{"sc"},
				Usage:   "Copy all your change to the plugin repo",
				Action: func(c *cli.Context) (err error) {
					stageChanges()
					return
				},
			},
			{
				Name:      "update-version",
				Aliases:   []string{"uv"},
				Usage:     "Updated the version number in package.json and files specified by --files flag",
				ArgsUsage: "[major|minor|patch]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "files",
						Aliases:     []string{"f"},
						Usage:       "Add a comma seperated list of files you want to change version numbers in",
						Destination: &files,
					},
				},
				Action: func(c *cli.Context) (err error) {
					level := c.Args().First()
					// Check if level is valid
					if level != "major" && level != "minor" && level != "patch" {
						fmt.Println(
							"Invalid level. Use one of the following: major, minor, patch.",
						)
						return
					}

					updateVersionNumbers(level, files)
					return
				},
			},
			{
				Name:  "setup",
				Usage: "Setup needed dirs and stuff for dc-plugged",
				Action: func(c *cli.Context) (err error) {
					// Check if ~/plugins exists, if not create it
					_, err = os.Stat(pluginsDir)
					if os.IsNotExist(err) {
						fmt.Println("~/plugins/ does not exist. Creating it.")
						err = os.Mkdir(pluginsDir, 0755)
						check(err)
					}

					fmt.Println("Setup complete.")
					return
				},
			},
		},
	}

	err := app.Run(os.Args)
	check(err)
}
