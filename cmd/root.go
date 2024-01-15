package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	pkg "github.com/cmgriffing/will-it-blend/pkg"
	"github.com/google/shlex"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Constants for default configuration values
const (
    DefaultCfgFile       = "~/.config/.will-it-blend.yaml"
    DefaultTitle         = "Will it blend?"
    DefaultDuration      = 60
    DefaultSuccessMsg    = "Yes"
    DefaultFailureMsg    = "No"
    DefaultAuthToken     = ""
)

// Command-line flags
var (
    cfgFile       string
    title         string
    duration      int
    successMsg    string
    failureMsg    string
    port          pkg.AllowedPort
    token         string
    successMessage string
    failureMessage string
)

// RootCmd is the root command for the will-it-blend CLI application
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

func Init() {
    port = pkg.AllowedPort3000
    cobra.OnInitialize(initConfig)
    initializeFlags()
}

// initializeFlags sets up the command line flags
func initializeFlags() {
    RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", DefaultCfgFile, "config file (default is $HOME/.config/.will-it-blend.yaml)")
    RootCmd.PersistentFlags().StringVarP(&title, "title", "t", "Will it blend?", "title of prediction (default is \"Will it blend?\")")
	RootCmd.PersistentFlags().IntVarP(&duration, "duration", "d", 60, "the duration of the prediction timer (default is 60 seconds)")
	RootCmd.PersistentFlags().StringVarP(&successMessage, "success", "s", "Yes", "success message of prediction (default is \"Yes\")")
	RootCmd.PersistentFlags().StringVarP(&failureMessage, "failure", "f", "No", "failure message of prediction (default is \"No\")")
	RootCmd.PersistentFlags().StringVarP(&token, "token", "k", "", "auth token for Twitch")
	RootCmd.Flags().VarP(&port, "port", "p", `auth server port. allowed: 1337, 3000, 4242, 6666, 6969, 8000, 8008, 8080, 42069`)
    viper.BindPFlag("title", RootCmd.PersistentFlags().Lookup("title"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
    viper.AutomaticEnv() // read in environment variables that match

    if cfgFile != "" {
        viper.SetConfigFile(cfgFile)
    } else {
        // Find home directory and set the default config path
        home, err := homedir.Dir()
        if err != nil {
            log.Fatalf("Failed to find homedir: %s", err)
        }
        viper.AddConfigPath(fmt.Sprintf("%s/.config", home))
        viper.SetConfigName(".will-it-blend")
    }

    // If a config file is found, read it in.
    if err := viper.ReadInConfig(); err == nil {
        log.Printf("Using config file: %s", viper.ConfigFileUsed())
    }
}

// runCommandPrediction automates the creation of CLI command based Twitch.tv predictions
func runCommandPrediction(cmd *cobra.Command, args []string) {
    command := args[0]

    if token == "" {
        token = pkg.StartServer(port)
    }

    userId, err := pkg.GetUserId(token)
    if err != nil {
        log.Fatalf("Error getting user ID: %s", err)
    }
    
    prediction, err := createAndRunPrediction(token, userId)
    if err != nil {
        log.Fatalf("Error running prediction: %s", err)
    }

    if !awaitPredictionCompletion(token, userId, prediction.PredictionId) {
        log.Fatal("Prediction did not lock as expected")
    }

    resolvePredictionBasedOnCommand(command, token, userId, prediction)
}

// createAndRunPrediction creates a new prediction and waits for its duration
func createAndRunPrediction(token string, userId string) (pkg.CreatePredictionResult, error) {
    prediction, err := pkg.CreatePrediction(token, title, userId, successMsg, failureMsg, duration)
    if err != nil {
        return pkg.CreatePredictionResult{}, fmt.Errorf("failed to create prediction: %w", err)
    }
    fmt.Printf("Prediction running. Command will run after %d seconds\n", duration)
    time.Sleep(time.Duration(duration) * time.Second)
    return prediction, nil
}

// awaitPredictionCompletion checks if the prediction has finished
func awaitPredictionCompletion(token, userId, predictionId string) bool {
    const loopSleep = 5
    const loopMax = 18 // 90 seconds divided by 5
    for loopCount := 0; loopCount < loopMax; loopCount++ {
        time.Sleep(loopSleep * time.Second)
        finished, err := pkg.IsPredictionFinished(token, userId, predictionId)
        if err != nil {
            log.Fatalf("Failed to check if prediction is finished: %v", err)
        }
        if finished {
            return true
        }
    }
    return false
}

// resolvePredictionBasedOnCommand executes the command and resolves the prediction based on its success or failure
func resolvePredictionBasedOnCommand(command, token, userId string, prediction pkg.CreatePredictionResult) {
    commandSuccess := runCommand(command)

    winningOutcomeId := prediction.SuccessOutcomeId
    if !commandSuccess {
        winningOutcomeId = prediction.FailureOutcomeId
    }

    pkg.ResolvePrediction(token, userId, prediction.PredictionId, winningOutcomeId)
    log.Println("Prediction Resolved")
}

// runCommand executes the given shell command and returns true if it succeeds
func runCommand(command string) bool {
    parts, err := shlex.Split(command)
    if err != nil {
        log.Printf("Failed to parse command: %s", err)
        return false
    }

    cmd := exec.Command(parts[0], parts[1:]...)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    if err := cmd.Start(); err != nil {
        log.Printf("Failed to start command: %s", err)
        return false
    }

    if err := cmd.Wait(); err != nil {
        log.Printf("Command execution failed: %s", err)
        return false
    }

    return true
}
