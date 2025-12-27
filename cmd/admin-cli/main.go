package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

// admin-cli 是一个独立的命令行工具，用于带外(Out-of-Band)管理操作
// 必须通过 SSH 登录服务器并在本地 shell 执行，严禁通过 Web 触发

func main() {
	resetCmd := flag.NewFlagSet("reset-passkey", flag.ExitOnError)
	username := resetCmd.String("user", "admin", "Target username to reset")
	force := resetCmd.Bool("force", false, "Skip confirmation")

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "reset-passkey":
		resetCmd.Parse(os.Args[2:])
		handleResetPasskey(*username, *force)
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Alpha-Trade Admin CLI Tool")
	fmt.Println("Usage:")
	fmt.Println("  admin-cli reset-passkey --user=<username> [--force]")
}

func handleResetPasskey(user string, force bool) {
	// TODO: Connect to DB
	// TODO: Verify admin password (interactive input)
	// TODO: Delete rows from webauthn_credentials where user_id = ?
	// TODO: Generate and print magic link token
	fmt.Printf("[MOCK] Resetting passkeys for user: %s\n", user)
	if !force {
		fmt.Print("Are you sure? This will delete ALL passkeys for this user. (y/N): ")
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "y" {
			fmt.Println("Aborted.")
			return
		}
	}
	fmt.Println("Success. Passkeys deleted.")
	fmt.Println("Magic Link: https://api.alpha-trade.internal/auth/magic-register?token=xyz123")
}

