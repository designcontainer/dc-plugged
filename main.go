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

var pluginDir string = path.Join(getHomeDir(), "plugins")

func getCWD() (dir string) {
	dir, err := os.Getwd()
	check(err)

	return
}

func getThePluginDir() (dir string) {
	dir = path.Base(getCWD())
	dir = path.Join(pluginDir, dir)

	return
}

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

// DONE
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
		fmt.Println("Repo is dirty, please commit or stash your changes before creating a new branch.")
		return
	}

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
		fmt.Printf("Created branch `%s` in `%s`\n", branchName, getThePluginDir())
	}
}

func stageChanges() {
	ignoredFiles := []string{".git", ".gitignore", "node_modules", ".env"}

	/**
	 * Delete all files in the plugin directory that are not in the ignoredFiles list
	 */
	files, err := os.ReadDir(getThePluginDir())
	for _, dirEntry := range files {
		if !contains(ignoredFiles, dirEntry.Name()) {
			os.RemoveAll(path.Join(getThePluginDir(), dirEntry.Name()))
		}
	}

	/**
	 * Copy over files from CWD to the plugin directory
	 */
	files, err = os.ReadDir(getCWD())
	check(err)
	for _, dirEntry := range files {
		// Copy file to the pluginDir if the filename is not in the ignoredFiles array
		if !contains(ignoredFiles, dirEntry.Name()) {
			var cp *exec.Cmd

			if dirEntry.IsDir() {
				cp = exec.Command(
					"cp",
					"-R",
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

	fmt.Println("Staged changes. They are ready to be commited in " + getThePluginDir())
}

func updateVersionNumbers(level string) {
	// Check if level is valid
	if level != "major" && level != "minor" && level != "patch" {
		fmt.Println("Invalid level, please use one of the following: major, minor, patch.")
		return
	}

	// Read the package.json file
	packageJSONBytes, err := os.ReadFile(
		path.Join(getThePluginDir(), "package.json"),
	)
	check(err)

	// Unmarshal the json into a map
	var packageJSONMap map[string]interface{}
	json.Unmarshal(packageJSONBytes, &packageJSONMap)

	// Increment the version number according to the level
	version := strings.Split(packageJSONMap["version"].(string), ".")
	var num int

	// Get the index of the version number to increment
	var table = map[string]int{
		"major": 0,
		"minor": 1,
		"patch": 2,
	}
	index := table[level]

	// Update the version number
	num, err = strconv.Atoi(version[index])
	version[index] = fmt.Sprint(num + 1)

	check(err)

	// Upgrade notice
	fmt.Printf(
		"Updated verson from %s to %s.\n",
		packageJSONMap["version"],
		strings.Join(version, "."),
	)

	// Replace the version number in the package.json file
	fileString := strings.Replace(
		string(packageJSONBytes),
		packageJSONMap["version"].(string),
		strings.Join( // Join them back to a string
			version,
			".",
		),
		1, // Only replace the first instance
	)

	// Save the new version number to the package.json file
	os.WriteFile(
		path.Join(getThePluginDir(), "package.json"),
		[]byte(fileString),
		0644,
	)
}

func main() {
	app := &cli.App{
		Name:    "DC-Plugged",
		Usage:   "Make it easier to work with plugins in site repos",
		Version: "0.1.0",
		Commands: []*cli.Command{
			{
				Name:        "new-branch",
				Aliases:     []string{"nb"},
				ArgsUsage:   "<branchName>",
				Description: "Create a new branch in the plugin you are currently editing.",
				Action: func(c *cli.Context) (err error) {
					branchName := c.Args().First()

					if branchName == "" {
						fmt.Println("Please provide a branch name.")
						return
					}

					newBranch(branchName)
					return
				},
			},
			{
				Name:        "stage-changes",
				Aliases:     []string{"sc"},
				Description: "Add all your changes to the plugin in the site repo to the plugin repo",
				Action: func(c *cli.Context) (err error) {
					stageChanges()
					return
				},
			},
			{
				Name:        "update-version",
				Aliases:     []string{"uv"},
				Usage:       "<level [major|minor|patch]>",
				UsageText:   "",
				Description: "",
				ArgsUsage:   "",
				Action: func(c *cli.Context) (err error) {
					level := c.Args().First()
					updateVersionNumbers(level)
					return
				},
			},
			{
				Name:  "setup",
				Usage: "Setup needed dirs and stuff for dc-plugged",
				Action: func(c *cli.Context) (err error) {
					// Check if ~/plugins exists, if not create it
					_, err = os.Stat(pluginDir)
					if os.IsNotExist(err) {
						err = os.Mkdir(pluginDir, 0755)
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
