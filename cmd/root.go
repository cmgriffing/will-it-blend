package cmd

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"time"

	pkg "github.com/cmgriffing/will-it-blend/pkg"
	"github.com/google/shlex"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	defaultCfgFile       = "~/.config/.will-it-blend.yaml"
	defaultTitle         = "Will it blend?"
	defaultDuration      = 60
	defaultSuccessMsg    = "Yes"
	defaultFailureMsg    = "No"
	defaultAuthToken     = ""
)

var (
	cfgFile       string
	title         string
	duration      int
	successMsg    string
	failureMsg    string
	port          pkg.PortFlag
	token         string
)

var RootCmd = &cobra.Command{
	Use:     "will-it-blend",
	Version: "0.3.2",
	Short:   "will-it-blend is a tool for automating the creation of CLI command based Twitch.tv predictions",
	Long: `will-it-blend is a tool for automating the creation of CLI command based Twitch.tv predictions.

It will resolve the prediction based on the return code of the specified command.`,
	Example: "will-it-blend \"ls -a\"",
	Args:    cobra.ExactArgs(1),
	Run:     runCommandPrediction,
}

func init() {
	RootCmd.Flags().SortFlags = false
	RootCmd.PersistentFlags().SortFlags = false

	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", defaultCfgFile, "config file (default is $HOME/.config/.will-it-blend.yaml)")
	RootCmd.PersistentFlags().StringVarP(&title, "title", "t", defaultTitle, "title of prediction (default is \"Will it blend?\")")
	RootCmd.PersistentFlags().IntVarP(&duration, "duration", "d", defaultDuration, "the duration of the prediction timer (default is 60 seconds)")
	RootCmd.PersistentFlags().StringVarP(&successMsg, "success", "s", defaultSuccessMsg, "success message of prediction (default is \"Yes\")")
	RootCmd.PersistentFlags().StringVarP(&failureMsg, "failure", "f", defaultFailureMsg, "failure message of prediction (default is \"No\")")

	RootCmd.PersistentFlags().StringVarP(&token, "token", "k", defaultAuthToken, "auth token for Twitch")

	RootCmd.Flags().VarP(&port, "port", "p", `auth server port. allowed: 1337, 3000, 4242, 6666, 6969, 8000, 8008, 8080, 42069`)

	cobra.OnInitialize(initConfig)

	viper.BindPFlag("title", RootCmd.PersistentFlags().Lookup("title"))
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

func runCommandPrediction(cmd *cobra.Command, args []string) {
	command := args[0]

	if token == "" {
		token = pkg.StartServer(port)
	}

	userId := pkg.GetUserId(token)

	prediction := pkg.CreatePrediction(token, title, userId, successMsg, failureMsg, duration)

	fmt.Printf("Prediction running. Command will run after %d seconds\n", duration)

	time.Sleep(time.Duration(duration) * time.Second)

	loopCount := 0
	loopSleep := 5
	loopMax := 90 / loopSleep
	finished := false

	for loopCount < loopMax {
		time.Sleep(time.Duration(loopSleep) * time.Second)
		finished = pkg.IsPredictionFinished(token, userId, prediction.PredictionId)
		loopCount++
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
}

func runCommand(command string) bool {
	parts, err := shlex.Split(command)

	if err != nil {
		fmt.Println("Failed to parse command", err)
		return false
	}

	cmd := exec.Command(parts[0], parts[1:]...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Start()

	if err != nil {
		fmt.Println("Failed to start command")
		return false
	}

	err = cmd.Wait()

	if err != nil {
		return false
	}

	return true
}
