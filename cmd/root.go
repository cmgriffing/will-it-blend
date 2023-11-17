package cmd

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	pkg "github.com/cmgriffing/will-it-blend/pkg"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var wg sync.WaitGroup

var cfgFile = ""
var title = ""
var duration = 30
var successMessage = ""
var failureMessage = ""
var port = pkg.AllowedPort3000
var token = ""

var RootCmd = &cobra.Command{
	Use:     "will-it-blend",
	Version: "0.0.1",
	Short:   "will-it-blend is a tool for automating the creation of CLI command based Twitch.tv predictions",
	Long: `will-it-blend is a tool for automating the creation of CLI command based Twitch.tv predictions.

It will resolve the prediction based on the return code of the specified command.`,
	Example: "will-it-blend \"ls -a\"",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		command := args[0]
		// fmt.Println("cfgFile", cfgFile)
		// fmt.Println("command in rootCmd", command)
		// fmt.Println("title", title)
		// fmt.Println("duration", duration)
		// fmt.Println("successMessage", successMessage)
		// fmt.Println("failureMessage", failureMessage)
		// fmt.Println("port", port)
		// fmt.Println("token", token)

		if token == "" {
			token = pkg.StartServer(port)
		}

		userId := pkg.GetUserId(token)

		prediction := pkg.CreatePrediction(token, title, userId, successMessage, failureMessage, duration)

		fmt.Printf("Prediction running. Command will run after %d seconds \n", duration)

		time.Sleep(time.Duration(duration) * time.Second)

		loopCount := 0
		loopSleep := 5
		loopMax := 90 / loopSleep
		finished := false

		for loopCount < loopMax {
			time.Sleep(time.Duration(loopSleep) * time.Second)
			finished = pkg.IsPredictionFinished(token, userId, prediction.PredictionId)
			loopCount = loopCount + 1
			if finished {
				loopCount = math.MaxInt
			}
		}

		if !finished {
			fmt.Println("Prediction did not lock as expected")
			os.Exit(1)
		}

		commandSuccess := runCommand(command)

		winningOutcomeId := prediction.SuccessOutcomeId
		if !commandSuccess {
			winningOutcomeId = prediction.FailureOutcomeId
		}

		pkg.ResolvePrediction(token, userId, prediction.PredictionId, winningOutcomeId)

		fmt.Println("Prediction Resolved")

	},
}

func Init() {

	RootCmd.Flags().SortFlags = false
	RootCmd.PersistentFlags().SortFlags = false

	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "~/.config/.will-it-blend.yaml", "config file (default is $HOME/.config/.will-it-blend.yaml)")
	RootCmd.PersistentFlags().StringVarP(&title, "title", "t", "Will it blend?", "title of prediction (default is \"Will it blend?\")")
	RootCmd.PersistentFlags().IntVarP(&duration, "duration", "d", 30, "the duration of the prediction timer (default is 30 seconds)")
	RootCmd.PersistentFlags().StringVarP(&successMessage, "success", "s", "Yes", "success message of prediction (default is \"Yes\")")
	RootCmd.PersistentFlags().StringVarP(&failureMessage, "failure", "f", "No", "failure message of prediction (default is \"No\")")

	RootCmd.PersistentFlags().StringVarP(&token, "token", "k", "", "auth token for Twitch")

	RootCmd.Flags().VarP(&port, "port", "p", `auth server port. allowed: "3000", "4242", "6969", "8000", "8008", "8080", "42069"`)

	cobra.OnInitialize(initConfig)

	viper.BindPFlag("title", RootCmd.PersistentFlags().Lookup("title"))
	allSettings := viper.AllSettings()
	if allSettings != nil {
		fmt.Println("allSettings", allSettings)
		fmt.Println("title", viper.GetString("title"))
	}
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
		viper.WriteConfig()
	} else {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println("Failed to find homedir", err)
			os.Exit(1)
		}

		configDir := fmt.Sprintf("%s/.config", home)

		configDirExists, err := os.Stat(configDir)
		if err != nil && !os.IsNotExist(err) {
			fmt.Println(err)
			os.Exit(1)
		}
		if !configDirExists.IsDir() {
			fmt.Println("HOME/.config exists but is not a directory")
			os.Exit(1)
		}

		viper.AddConfigPath(configDir)
		viper.SetConfigName(".will-it-blend")
	}

}

func runCommand(command string) bool {
	parts := strings.Split(command, " ")

	cmd := exec.Command(parts[0], parts[1:]...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	defer cmd.Wait()

	if err != nil {
		return false
	} else {
		return true
	}
}
