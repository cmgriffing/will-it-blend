package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
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

func init() {
    cobra.OnInitialize(initConfig)
    initializeFlags()
}

// initializeFlags sets up the command line flags
func initializeFlags() {
    RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", DefaultCfgFile, "config file (default is $HOME/.config/.will-it-blend.yaml)")
    // Other flag initializations...
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

    userIdStr := strconv.Itoa(userId)
    prediction, err := createAndRunPrediction(token, userIdStr)
    if err != nil {
        log.Fatalf("Error running prediction: %s", err)
    }

    if !awaitPredictionCompletion(token, userIdStr, prediction.PredictionId) {
        log.Fatal("Prediction did not lock as expected")
    }

    resolvePredictionBasedOnCommand(command, token, userIdStr, prediction)
}

// createAndRunPrediction creates a new prediction and waits for its duration
func createAndRunPrediction(token string, userId string) (pkg.CreatePredictionResult, error) {
    prediction := pkg.CreatePrediction(token, title, userId, successMsg, failureMsg, duration)
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
        if finished := pkg.IsPredictionFinished(token, userId, predictionId); finished {
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
