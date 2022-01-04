package main

import (
	"bytes"
        "context"
        "encoding/json"
        "fmt"
        "io/ioutil"
	jpeg "image/jpeg"
        "log"
	"math/rand"
        "net/http"
        "os"
	"strconv"

        "golang.org/x/oauth2"
        "golang.org/x/oauth2/google"
	gphotos "github.com/gphotosuploader/google-photos-api-client-go/v2"
        resize "github.com/nfnt/resize"
	webp "github.com/chai2010/webp"
)

type Config struct {
	AlbumID   string `json:"albumId"`
	CredsFile string `json:"credsFile"`
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
        // The file token.json stores the user's access and refresh tokens, and is
        // created automatically when the authorization flow completes for the first
        // time.
        tokFile := "token.json"
        tok, err := tokenFromFile(tokFile)
        if err != nil {
                tok = getTokenFromWeb(config)
                saveToken(tokFile, tok)
        }
        return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
        authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
        fmt.Printf("Go to the following link in your browser then type the "+
                "authorization code: \n%v\n", authURL)

        var authCode string
        if _, err := fmt.Scan(&authCode); err != nil {
                log.Fatalf("Unable to read authorization code: %v", err)
        }

        tok, err := config.Exchange(context.TODO(), authCode)
        if err != nil {
                log.Fatalf("Unable to retrieve token from web: %v", err)
        }
        return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
        f, err := os.Open(file)
        if err != nil {
                return nil, err
        }
        defer f.Close()
        tok := &oauth2.Token{}
        err = json.NewDecoder(f).Decode(tok)
        return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
        fmt.Printf("Saving credential file to: %s\n", path)
        f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
        if err != nil {
                log.Fatalf("Unable to cache oauth token: %v", err)
        }
        defer f.Close()
        json.NewEncoder(f).Encode(token)
}

func readConfig() (string, string, error) {
	jsonFile, err := os.Open("latestphototidbyt.config.json")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Successfully opened configuration file...")
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var config Config
	json.Unmarshal(byteValue, &config)
	return config.CredsFile, config.AlbumID, nil
//	return "/etc/tidbytrandomphoto.json", "ABhkwBQRMsVffW1pB9ObuZ9XK8Y4cAeKktFrE2YVlNQDP4NSDdEatR8ihm1MvkcrN_J4s8M-7I5q", nil
}

func main() {
	credsFile, albumId, err := readConfig()
	var buf bytes.Buffer
        ctx := context.Background()
        b, err := ioutil.ReadFile(credsFile)
        if err != nil {
                log.Fatalf("Unable to read client secret file: %v", err)
        }

        // If modifying these scopes, delete your previously saved token.json.
        config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/photoslibrary")
        if err != nil {
                log.Fatalf("Unable to parse client secret file to config: %v", err)
        }
        client := getClient(config)
	gpc, err := gphotos.NewClient(client)
	ta, err := gpc.Albums.GetById(ctx, albumId)
	if err != nil {
                log.Fatal(err)
        }
	taid := ta.ID
        log.Println("Album ID is: " + ta.ID)
	log.Println("Album Name is: " + ta.Title)
	log.Println("Album Count is: " + ta.MediaItemsCount)
	mic,err := strconv.Atoi(ta.MediaItemsCount)
	if err != nil {
		log.Fatal(err)
	}
	cmi := rand.Intn(mic - 1)
	log.Println("Chosen Media Index is: " + strconv.Itoa(cmi))

	am, err := gpc.MediaItems.ListByAlbum(ctx, taid)
	log.Println("Number of Media Items Returned is: " + strconv.Itoa(len(am)))
	mediaUrl := ""
	for _, mi := range am {
		if mi.MediaMetadata.Width > mi.MediaMetadata.Height {
	                log.Println("ID: " + mi.ID)
			log.Println("Creation Time: " + mi.MediaMetadata.CreationTime)
	                log.Println("--------------------------")
			mediaUrl = mi.BaseURL + "=w64"
			break
		}
        }

        resp, err := http.Get(mediaUrl)
        if err != nil {
                log.Println(err)
        }
        defer resp.Body.Close()
        imgData := resp.Body

	// Decode jpg
        image, err := jpeg.Decode(imgData)
        if err != nil {
                log.Println(err)
        }
        resizedImage := resize.Resize(64, 32, image, resize.Lanczos3)
        // Encode lossless webp
        if err = webp.Encode(&buf, resizedImage, &webp.Options{Lossless: true}); err != nil {
                log.Println(err)
        }
        if err = ioutil.WriteFile("output.webp", buf.Bytes(), 0666); err != nil {
                log.Println(err)
        }

	fmt.Println("Save output.webp ok")
}
