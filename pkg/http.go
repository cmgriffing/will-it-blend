package pkg

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"text/template"

	"github.com/spf13/viper"
)

var wg sync.WaitGroup

var Token = ""

func StartServer(port AllowedPort) string {
	http.HandleFunc("/", createIndexHandler(port))
	http.HandleFunc("/auth", createAuthHandler(port))
	wg.Add(1)
	go func() {
		err := http.ListenAndServe(fmt.Sprintf("localhost:%s", port), nil)
		if err != nil {
			fmt.Println("Could not start auth server at: ", port, err)
			os.Exit(1)
		}
	}()
	fmt.Printf("Auth server started at http://localhost:%s\n", port)
	openBrowser(fmt.Sprintf("https://id.twitch.tv/oauth2/authorize?client_id=%s&redirect_uri=http://localhost:%s&response_type=token&scope=channel:manage:predictions", CLIENT_ID, port))

	wg.Wait()
	return Token
}

type IndexValues struct {
	Port     AllowedPort
	Scope    string
	ClientId string
}

func createIndexHandler(port AllowedPort) func(response http.ResponseWriter, request *http.Request) {

	return func(response http.ResponseWriter, request *http.Request) {
		var parsedTemplate, err = template.New("index.html").Parse(indexTemplate)

		if err != nil {
			fmt.Println("Could not parse index.html template")
			os.Exit(1)
		}

		var executeErr = parsedTemplate.Execute(response, IndexValues{
			Port:     port,
			Scope:    "channel:manage:predictions",
			ClientId: CLIENT_ID,
		})

		if executeErr != nil {
			fmt.Println("Could not execute index.html template with provided data", executeErr)
			os.Exit(1)
		}
	}
}

func createAuthHandler(port AllowedPort) func(response http.ResponseWriter, request *http.Request) {
	return func(response http.ResponseWriter, request *http.Request) {

		var postAuthBody PostAuthBody
		err := json.NewDecoder(request.Body).Decode(&postAuthBody)
		if err != nil {
			fmt.Println("Could not decode request body at POST:Auth")
			os.Exit(1)
		}

		viper.Set("token", postAuthBody.Token)
		Token = postAuthBody.Token

		wg.Done()
	}
}

type PostAuthBody struct {
	Token string `json:"token"`
}

func openBrowser(url string) {

	fmt.Println("Opening browser for Twitch OAuth flow.")

	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}

type AllowedPort string

const (
	AllowedPort3000  AllowedPort = "3000"
	AllowedPort4242  AllowedPort = "4242"
	AllowedPort6969  AllowedPort = "6969"
	AllowedPort8000  AllowedPort = "8000"
	AllowedPort8008  AllowedPort = "8008"
	AllowedPort8080  AllowedPort = "8080"
	AllowedPort42069 AllowedPort = "42069"
)

// String is used both by fmt.Print and by Cobra in help text
func (e *AllowedPort) String() string {
	return string(*e)
}

// Set must have pointer receiver so it doesn't change the value of a copy
func (e *AllowedPort) Set(v string) error {
	switch v {
	case "1337", "3000", "4242", "6666", "6969", "8000", "8008", "8080", "42069":
		*e = AllowedPort(v)
		return nil
	default:
		return errors.New(`must be one of 1337, 3000, 4242, 6666, 6969, 8000, 8008, 8080, or 42069`)
	}
}

// Type is only used in help text
func (e *AllowedPort) Type() string {
	return "AllowedPort"
}

var indexTemplate = `
<html>

<head>
  <style>
    h1 {
      margin: 4rem;
    }

    body, main {
      text-align: center;
    }

    a {
      color: white;
      background-color: #9147FF;
      text-decoration: none;
      padding: 1rem;
    }

    .error {
      display: none;
    }
  </style>
</head>

<body>

  <h1>Will It Blend?</h1>

  <noscript>

    <h2>Error: JavaScript Required</h2>

    <p>This page requires JavaScript to be able to send a request to the local server.</p>

  </noscript>

  <main hidden>

    <a class="auth-button" href="https://id.twitch.tv/oauth2/authorize?client_id={{.ClientId}}&redirect_uri=http://localhost:{{.Port}}&response_type=token&scope={{.Scope}}">Login with Twitch</a>

    <div class="success" hidden>
      <h2>Successfully authorized via Twitch!</h2>
      <p>The command will run. You may now close this page.</p>
    </div>

    <div class="error" hidden>
      <h2>There has been an error logging into Twitch.</h2>
    </div>

  </main>

  <script type="text/javascript">

    document.querySelector('main').toggleAttribute("hidden");


    async function handleHash() {
      console.log("hash", location.hash)
      const accessToken = location.hash.split("&")?.[0].replace("#access_token=", "");

      if (!accessToken) {
        console.log("No access token. Bailing out")
        return;
      }

      const response = await fetch("http://localhost:{{.Port}}/auth", {
        method: 'POST',
        body: JSON.stringify({
          token: accessToken
        })
      });

      if (response.ok) {
        document.querySelector('.auth-button').toggleAttribute("hidden");
        document.querySelector('.success').toggleAttribute("hidden");
      } else {
        console.log("Error: Could not post auth.")
        document.querySelector('.error').toggleAttribute("hidden");
      }

  }

  handleHash();
  </script>


</body>
</html>
`
