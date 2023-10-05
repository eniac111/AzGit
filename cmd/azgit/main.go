package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
)

const azgitFolderName = ".config/azgit"
const azgitConfigFileName = "config.ini"

type GitIdentity struct {
	Name       string
	Email      string
	SigningKey string
	GPGSign    string
}

var rootCmd = &cobra.Command{
	Use:   "azgit",
	Short: "AzGit manages Git identities.",
	Long:  `A simple tool to manage and switch between different Git identities.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initializeAzGitConfig(); err != nil {
			return err
		}
		return cmd.Help()
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all identities.",
	RunE:  listIdentities,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func getAzgitConfigPath() string {
	return filepath.Join(userHomeDir(), azgitFolderName, azgitConfigFileName)
}

func loadAzGitConfig() (*ini.File, error) {
	return ini.Load(getAzgitConfigPath())
}

func ensureAzGitFolderExists() error {
	azgitFolderPath := filepath.Join(userHomeDir(), azgitFolderName)
	if _, err := os.Stat(azgitFolderPath); os.IsNotExist(err) {
		return os.MkdirAll(azgitFolderPath, 0755)
	}
	return nil
}

func fetchDefaultGitIdentity() (GitIdentity, error) {
	gitConfigPath := filepath.Join(userHomeDir(), ".gitconfig")
	gitCfg, err := ini.Load(gitConfigPath)
	if err != nil {
		return GitIdentity{}, fmt.Errorf("failed to read ~/.gitconfig: %w", err)
	}

	return GitIdentity{
		Name:       gitCfg.Section("user").Key("name").String(),
		Email:      gitCfg.Section("user").Key("email").String(),
		SigningKey: gitCfg.Section("user").Key("signingkey").String(),
		GPGSign:    gitCfg.Section("commit").Key("gpgsign").String(),
	}, nil
}

func initializeAzGitConfig() error {
	if _, err := os.Stat(getAzgitConfigPath()); !os.IsNotExist(err) {
		return nil
	}

	if err := ensureAzGitFolderExists(); err != nil {
		return fmt.Errorf("failed to ensure azgit config directory exists: %w", err)
	}

	identity, err := fetchDefaultGitIdentity()
	if err != nil {
		return err
	}

	newCfg := ini.Empty()
	defaultSection, err := newCfg.NewSection("default")
	if err != nil {
		return fmt.Errorf("failed to create default section in azgit config: %w", err)
	}

	defaultSection.NewKey("name", identity.Name)
	defaultSection.NewKey("email", identity.Email)
	if identity.SigningKey != "" {
		defaultSection.NewKey("signingkey", identity.SigningKey)
	}
	if identity.GPGSign != "" {
		defaultSection.NewKey("gpgsign", identity.GPGSign)
	}

	return newCfg.SaveTo(getAzgitConfigPath())
}

func listIdentities(cmd *cobra.Command, args []string) error {
	cfg, err := loadAzGitConfig()
	if err != nil {
		return fmt.Errorf("failed to read azgit config: %w", err)
	}

	fmt.Println("List of Identities:")
	for _, section := range cfg.Sections() {
		name := section.Key("name").String()
		email := section.Key("email").String()

		// Skip sections without valid data
		if name == "" && email == "" {
			continue
		}

		signingKey := section.Key("signingkey").String()
		gpgSign := section.Key("gpgsign").String()

		identityInfo := fmt.Sprintf("\nIdentity [%s]:\n", section.Name())
		identityInfo += fmt.Sprintf("\tName: %s\n", name)
		identityInfo += fmt.Sprintf("\tEmail: %s\n", email)
		if signingKey != "" {
			identityInfo += fmt.Sprintf("\tSigning Key: %s\n", signingKey)
		}
		if gpgSign != "" {
			identityInfo += fmt.Sprintf("\tGPG Sign: %s\n", gpgSign)
		}
		fmt.Println(identityInfo)
	}
	return nil
}

func userHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic("Failed to get user's home directory: " + err.Error())
	}
	return home
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
