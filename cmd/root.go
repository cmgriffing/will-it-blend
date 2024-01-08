package cmd

import (
    "fmt"
    "os"
    "os/exec"
    "time"

    "github.com/cmgriffing/will-it-blend/pkg"
    "github.com/google/shlex"
    "github.com/mitchellh/go-homedir"
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

// Config represents the configuration settings
type Config struct {
    File           string
    Title          string
    Duration       int
    SuccessMessage string
    FailureMessage string
    Port           pkg.AllowedPort
    Token          string
}

var cfg Config

var RootCmd = &cobra.Command{
    Use:     "will-it-blend",
    Version: "0.3.2",
    Short:   "will-it-blend is a tool for automating the creation of CLI command based Twitch.tv predictions",
    Long: `will-it-blend is a tool for automating the creation of CLI command based Twitch.tv predictions.
           It will resolve the prediction based on the return code of the specified command.`,
    Example: "will-it-blend \"ls -a\"",
    Args:    cobra.ExactArgs(1),
    Run:     runRootCmd,
}

func runRootCmd(cmd *cobra.Command, args []string) {
    command := args[0]

    // Token handling
    if cfg.Token == "" {
        cfg.Token = pkg.StartServer(cfg.Port)
    }

    userId := pkg.GetUserId(cfg.Token)
    prediction := pkg.CreatePrediction(cfg.Token, cfg.Title, userId, cfg.SuccessMessage, cfg.FailureMessage, cfg.Duration)

    fmt.Printf("Prediction running. Command will run after %d seconds\n", cfg.Duration)
    time.Sleep(time.Duration(cfg.Duration) * time.Second)

    loopCount := 0
    loopSleep := 5
    loopMax := 90 / loopSleep
    finished := false

    for loopCount < loopMax && !finished {
        time.Sleep(time.Duration(loopSleep) * time.Second)
        finished = pkg.IsPredictionFinished(cfg.Token, userId, prediction.PredictionId)
        loopCount++
    }

    if !finished {
        fmt.Println("Prediction did not lock as expected")
        return // Return error in real-world usage
    }

    commandSuccess := runCommand(command)
    winningOutcomeId := getWinningOutcomeId(prediction, commandSuccess)
    pkg.ResolvePrediction(cfg.Token, userId, prediction.PredictionId, winningOutcomeId)

    fmt.Println("Prediction Resolved")
}

func getWinningOutcomeId(prediction pkg.Prediction, commandSuccess bool) string {
    if commandSuccess {
        return prediction.SuccessOutcomeId
    }
    return prediction.FailureOutcomeId
}

func runCommand(command string) bool {
    parts, err := shlex.Split(command)
    if err != nil {
        fmt.Println("Failed to parse command:", err)
        return false
    }

    cmd := exec.Command(parts[0], parts[1:]...)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    if err := cmd.Start(); err != nil {
        fmt.Println("Failed to start command:", err)
        return false
    }

    if err := cmd.Wait(); err != nil {
        return false
    }
    return true
}

func Init() {
    RootCmd.PersistentFlags().StringVarP(&cfg.File, "config", "c", "~/.config/.will-it-blend.yaml", "config file (default is $HOME/.config/.will-it-blend.yaml)")
    RootCmd.PersistentFlags().StringVarP(&cfg.Title, "title", "t", "Will it blend?", "title of prediction (default is \"Will it blend?\")")
    RootCmd.PersistentFlags().IntVarP(&cfg.Duration, "duration", "d", 60, "the duration of the prediction timer (default is 60 seconds)")
    RootCmd.PersistentFlags().StringVarP(&cfg.SuccessMessage, "success", "s", "Yes", "success message of prediction (default is \"Yes\")")
    RootCmd.PersistentFlags().StringVarP(&cfg.FailureMessage, "failure", "f", "No", "failure message of prediction (default is \"No\")")
    RootCmd.PersistentFlags().StringVarP(&cfg.Token, "token", "k", "", "auth token for Twitch")
    RootCmd.Flags().VarP(&cfg.Port, "port", "p", `auth server port. allowed: 1337, 3000, 4242, 6666, 6969, 8000, 8008, 8080, 42069`)

    cobra.OnInitialize(initConfig)
}

func initConfig() {
    if cfg.File != "" {
        viper.SetConfigFile(cfg.File)
        viper.WriteConfig()
    } else {
        home, err := homedir.Dir()
        if err != nil {
            fmt.Println("Failed to find homedir", err)
            return
        }

        configDir := fmt.Sprintf("%s/.config", home)

        configDirExists, err := os.Stat(configDir)
        if os.IsNotExist(err) {
            fmt.Println("HOME/.config does not exist")
            return
        }
        if err != nil {
            fmt.Println(err)
            return
        }
        if !configDirExists.IsDir() {
            fmt.Println("HOME/.config exists but is not a directory")
            return
        }

        viper.AddConfigPath(configDir)
        viper.SetConfigName(".will-it-blend")
    }
}
