package main

import (
	"IndicoDataFusion/backend/utils"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "set":
		handleSet()
	case "get":
		handleGet()
	case "delete":
		handleDelete()
	case "test":
		handleTest()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: manage-secrets <command> [arguments]")
	fmt.Println("\nCommands:")
	fmt.Println("  set <name> <token>   Store a token securely")
	fmt.Println("  get <name>           Retrieve a stored token")
	fmt.Println("  delete <name>        Delete a stored token")
	fmt.Println("  test                 Check keyring backend by doing a temporary set/get/delete")
}

func handleSet() {
	setCmd := flag.NewFlagSet("set", flag.ExitOnError)
	if err := setCmd.Parse(os.Args[2:]); err != nil {
		fmt.Printf("Failed to parse 'set' arguments: %v\n", err)
		os.Exit(2)
	}

	args := setCmd.Args()
	if len(args) != 2 {
		fmt.Println("Usage: manage-secrets set <name> <token>")
		os.Exit(1)
	}

	name := args[0]
	token := args[1]

	err := utils.SetAPITokenSecret(name, token)
	if err != nil {
		fmt.Printf("Error setting token: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Token '%s' saved successfully.\n", name)
}

func handleGet() {
	getCmd := flag.NewFlagSet("get", flag.ExitOnError)
	if err := getCmd.Parse(os.Args[2:]); err != nil {
		fmt.Printf("Failed to parse 'get' arguments: %v\n", err)
		os.Exit(2)
	}

	args := getCmd.Args()
	if len(args) != 1 {
		fmt.Println("Usage: manage-secrets get <name>")
		os.Exit(1)
	}

	name := args[0]

	token, err := utils.GetAPITokenSecret(name)
	if err != nil {
		if strings.Contains(err.Error(), "secret not found") || strings.Contains(err.Error(), "item not found") {
			fmt.Printf("Token '%s' not found.\n", name)
		} else {
			fmt.Printf("Error getting token: %v\n", err)
		}
		os.Exit(1)
	}

	fmt.Println(token)
}

func handleDelete() {
	delCmd := flag.NewFlagSet("delete", flag.ExitOnError)
	if err := delCmd.Parse(os.Args[2:]); err != nil {
		fmt.Printf("Failed to parse 'delete' arguments: %v\n", err)
		os.Exit(2)
	}

	args := delCmd.Args()
	if len(args) != 1 {
		fmt.Println("Usage: manage-secrets delete <name>")
		os.Exit(1)
	}

	name := args[0]

	err := utils.DeleteAPITokenSecret(name)
	if err != nil {
		fmt.Printf("Error deleting token: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Token '%s' deleted successfully.\n", name)
}

func handleTest() {
	// Create a temporary unique key name for the test
	name := fmt.Sprintf("manage-secrets-test-%d-%d", os.Getpid(), time.Now().UnixNano())
	token := "test-value"

	fmt.Printf("Testing keyring backend using temporary key: %s\n", name)

	// Try to set
	if err := utils.SetAPITokenSecret(name, token); err != nil {
		printKeyringError("set", err)
		os.Exit(2)
	}

	// Try to get
	v, err := utils.GetAPITokenSecret(name)
	if err != nil {
		printKeyringError("get", err)
		// attempt cleanup if set partially
		_ = utils.DeleteAPITokenSecret(name)
		os.Exit(3)
	}

	if v != token {
		fmt.Printf("Unexpected value read back: %q\n", v)
		_ = utils.DeleteAPITokenSecret(name)
		os.Exit(4)
	}

	// Try to delete
	if err := utils.DeleteAPITokenSecret(name); err != nil {
		printKeyringError("delete", err)
		os.Exit(5)
	}

	fmt.Println("Keyring backend test succeeded.")
}

func printKeyringError(op string, err error) {
	// Print base error
	fmt.Printf("Keyring %s failed: %v\n", op, err)

	// Add actionable hints for common Linux issues
	errStr := strings.ToLower(err.Error())
	if strings.Contains(errStr, "secret service") || strings.Contains(errStr, "dbus") || strings.Contains(errStr, "gnome-keyring") || strings.Contains(errStr, "libsecret") {
		fmt.Println("Hints (Linux):")
		fmt.Println("  - Ensure a session secret service is available (gnome-keyring or libsecret).")
		fmt.Println("  - If running headless/CI, prefer providing tokens via environment variables or a file fallback.")
		fmt.Println("  - Check DBUS_SESSION_BUS_ADDRESS is set for the process or start gnome-keyring-daemon with secrets component.")
		fmt.Println("    Example: eval $(/usr/bin/gnome-keyring-daemon --start --components=secrets)")
		fmt.Println("  - On Debian/Ubuntu: sudo apt install gnome-keyring libsecret-tools")
	}
}
