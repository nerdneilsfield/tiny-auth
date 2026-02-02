package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
)

func newHashPasswordCmd() *cobra.Command {
	var bcryptCost int

	cmd := &cobra.Command{
		Use:   "hash-password [password]",
		Short: "Generate bcrypt hash for a password",
		Long: `Generate bcrypt hash for a password to use in configuration.

The generated hash can be used in the 'pass_hash' field of basic_auth configuration.

Example:
  tiny-auth hash-password "my-secret-password"

You can also specify the cost (default: 10, range: 4-31):
  tiny-auth hash-password "my-password" --cost 12

The output can be directly used in config.toml:
  pass_hash = "$2a$10$..."
`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			runHashPassword(cmd, args, bcryptCost)
		},
	}

	cmd.Flags().IntVar(&bcryptCost, "cost", 10, "bcrypt cost factor (4-31, default: 10)")

	return cmd
}

func runHashPassword(cmd *cobra.Command, args []string, bcryptCost int) {
	password := args[0]

	// 验证 cost 参数
	if bcryptCost < bcrypt.MinCost || bcryptCost > bcrypt.MaxCost {
		fmt.Printf("Error: cost must be between %d and %d\n", bcrypt.MinCost, bcrypt.MaxCost)
		return
	}

	// 生成 bcrypt 哈希
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		fmt.Printf("Error generating hash: %v\n", err)
		return
	}

	// 输出结果
	fmt.Println("\n✅ Bcrypt hash generated successfully!")
	fmt.Println("\n📋 Configuration:")
	fmt.Printf("pass_hash = \"%s\"\n", string(hash))
	fmt.Println("\n💡 Tips:")
	fmt.Println("  1. Copy the hash above to your config.toml")
	fmt.Println("  2. Remove or comment out the 'pass' field if using pass_hash")
	fmt.Println("  3. For environment variables: export PASSWORD_HASH=\"<hash>\"")
	fmt.Printf("     pass_hash = \"env:PASSWORD_HASH\"\n")
	fmt.Println("\n🔐 Security:")
	fmt.Printf("  - Cost factor: %d (higher = more secure but slower)\n", bcryptCost)
	fmt.Println("  - Never commit plain-text passwords to version control")
	fmt.Println("  - Store sensitive hashes in environment variables for production")
}
