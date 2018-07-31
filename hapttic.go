package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"unicode/utf8"

	minio "github.com/minio/minio-go"
)

const version = "1.0.0"

var minioEndpoint, minioAccessID, minioAccessSecret, minioBucket, minioLocation string
var minioSSL bool

func initMinio(endpoint string, accessKeyID string, secretAccessKey string, ssl bool) (minioClient *minio.Client, minioErr error) {
	// Initialize minio client object.
	client, minioErr := minio.New(endpoint, accessKeyID, secretAccessKey, ssl)
	if minioErr != nil {
		fmt.Println(minioErr)
	}
	return client, minioErr
}

func minioUpload(minioClient *minio.Client, bucketName string, location string, objectName string, filePath string) (minioErr error) {
	err := minioClient.MakeBucket(bucketName, location)
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, err := minioClient.BucketExists(bucketName)
		if err == nil && exists {
			log.Printf("We already own %s\n", bucketName)
		} else {
			log.Fatalln(err)
			return err
		}
	}
	log.Printf("Successfully created %s\n", bucketName)

	contentType := "application/db"

	// Upload the zip file with FPutObject
	n, err := minioClient.FPutObject(bucketName, objectName, filePath, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		log.Fatalln(err)
		return err
	}

	log.Printf("Successfully uploaded %s of size %d\n", objectName, n)
	return err
}

type minioDetails struct {
	minioEndpoint     string
	minioAccessID     string
	minioAccessSecret string
	minioLocation     string
	minioBucket       string
	minioSSL          bool
	useMinio          bool
}

// This is a subset of http.Request with the types changed so that we can marshall it.
type marshallableRequest struct {
	Method string
	URL    string
	Proto  string
	Host   string

	Header http.Header

	ContentLength int64
	Body          MapRequestBody
	Form          url.Values
	PostForm      url.Values
}

//MapRequestBody ... Struct for Mapbox Parameters
type MapRequestBody struct {
	NorthBound        string `json:"n-bound"`
	WestBound         string `json:"w-bound"`
	SouthBound        string `json:"s-bound"`
	EastBound         string `json:"e-bound"`
	MapStyle          string `json:"mapStyle"`
	MapName           string `json:"mapName"`
	MaxZoom           string `json:"maxZoom"`
	MinZoom           string `json:"minZoom"`
	PixelRatio        string `json:"pixelRatio"`
	OutputDir         string `json:"outputDir"`
	MapboxAccessToken string `json:"mapboxAccessToken"`
}

func init() {
	log.SetOutput(os.Stdout)
}

func ensureRequestHandlingScriptExists(scriptFileName string) {
	if _, err := os.Stat(scriptFileName); os.IsNotExist(err) {
		log.Fatal("The request handling script " + scriptFileName + " does not exist.")
	}
}

func ensureMinioFlagsExist(endpoint string, accessKeyID string, secretAccessKey string) (exists bool) {
	if endpoint == "" {
		flag.PrintDefaults()
		return false
	}
	if accessKeyID == "" {
		flag.PrintDefaults()
		return false
	}
	if secretAccessKey == "" {
		flag.PrintDefaults()
		return false
	}
	return true
}

// handleFuncWithScriptFileName constructs our handleFunc
func handleFuncWithScriptFileName(scriptFileName string, logErrorsToStderr bool, minioFlags minioDetails) func(s http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		ensureRequestHandlingScriptExists(scriptFileName)

		var mapRequest MapRequestBody
		err := json.NewDecoder(req.Body).Decode(&mapRequest)
		// r, err := json.Unmarshal(body, &mapRequest)
		if err != nil {
			fmt.Println("whoops:", err)
		}

		// req.ParseForm()

		// Try to convert to JSON. This shouldn't fail
		requestJSON, err := json.Marshal(mapRequest)
		if err != nil {
			log.Fatal(err)
		}
		cmdArgs := []string{"--north", mapRequest.NorthBound, "--west", mapRequest.WestBound, "--south", mapRequest.SouthBound, "--east", mapRequest.EastBound, "--output", mapRequest.MapName, "--style", mapRequest.MapStyle, "--token", mapRequest.MapboxAccessToken, "--minZoom", mapRequest.MinZoom, "--maxZoom", mapRequest.MaxZoom, "--pixelRatio", mapRequest.PixelRatio}
		log.Println("Arguments String")
		fmt.Printf("%v", cmdArgs)
		log.Println("Executing " + scriptFileName)
		var stdoutBuf, stderrBuf bytes.Buffer
		//Execute Command
		cmd := exec.Command(scriptFileName, cmdArgs...)
		//Pipe Progress From Execution to StdErr and StdOut
		stdoutIn, _ := cmd.StdoutPipe()
		stderrIn, _ := cmd.StderrPipe()
		var errStdout, errStderr error
		stdout := io.MultiWriter(os.Stdout, &stdoutBuf)
		stderr := io.MultiWriter(os.Stderr, &stderrBuf)
		cmdErr := cmd.Start()
		if cmdErr != nil {
			log.Fatalf("cmd.Start() failed with '%s'\n", cmdErr)
			// If there was an error, we return a response with status code 500
			res.WriteHeader(http.StatusInternalServerError)
			io.WriteString(res, "500 Internal Server Error: \n"+cmdErr.Error())
			if logErrorsToStderr {
				log.Println("\033[33;31m--- ERROR: ---\033[0m")
				log.Println("\033[33;31mParams:\033[0m")
				log.Println(string(requestJSON))
				log.Println("\033[33;31mScript output:\033[0m")
				log.Println(cmd)
				outStr, errStr := string(stdoutBuf.Bytes()), string(stderrBuf.Bytes())
				fmt.Printf("\nout:\n%s\nerr:\n%s\n", outStr, errStr)
				log.Println("\033[33;31m---- END: ----\033[0m")
			}
		}
		go func() {
			_, errStdout = io.Copy(stdout, stdoutIn)
		}()

		go func() {
			_, errStderr = io.Copy(stderr, stderrIn)
		}()
		err = cmd.Wait()
		if cmdErr != nil {
			log.Fatalf("cmd.Run() failed with %s\n", cmdErr)
		}
		if errStdout != nil || errStderr != nil {
			log.Fatal("failed to capture stdout or stderr\n")
		}
		outStr, errStr := string(stdoutBuf.Bytes()), string(stderrBuf.Bytes())
		fmt.Printf("\nStdout:\n%s\nStderr:\n%s\n", outStr, errStr)
		fmt.Println("Command Ran / Program Output: \n", cmd)
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
		res.Write(requestJSON)

		if minioFlags.useMinio == true {
			minioClient, err := initMinio(minioFlags.minioEndpoint, minioFlags.minioAccessID, minioFlags.minioAccessSecret, minioFlags.minioSSL)
			if err != nil {
				log.Fatalf("minio client initialization failed with %s\n", err)
				return
			}
			uploadErr := minioUpload(minioClient, minioFlags.minioBucket, minioFlags.minioLocation, mapRequest.MapName, mapRequest.OutputDir)
			if uploadErr != nil {
				log.Fatalf("minio client initialization failed with %s\n", uploadErr)
				return
			}
		}

	}
}

func main() {
	// Parse command line args
	printVersion := flag.Bool("version", false, "Print version and exit.")
	printUsage := flag.Bool("help", false, "Print usage and exit")
	host := flag.String("host", "", "The host to bind to, e.g. 0.0.0.0 or localhost.")
	port := flag.String("port", "8080", "The port to listen on.")
	userScriptFileName := flag.String("file", "./mbgl-offline", "The script that is called to handle requests.")
	logErrorsToStderr := flag.Bool("logErrors", false, "Log errors to stderr")
	minioEndpoint := flag.String("minioEndpoint", "", "minio Endpoint")
	minioAccessID := flag.String("minioAccessID", "", "minio Access Key ID")
	minioAccessSecret := flag.String("minioAccessSecret", "", "minio Access Key Secret")
	minioSSL := flag.Bool("minioSSL", false, "true/false use SSL for minio client")
	minioBucket := flag.String("minioBucket", "mapTiles", "Name of Minio Bucket to upload Cache tiles to")
	minioLocation := flag.String("minioLocation", "ship", "Minio Location Ship/Shore?")
	flag.Parse()

	fmt.Println("host:", *host)
	fmt.Println("port:", *port)
	fmt.Println("userScriptFileName:", *userScriptFileName)
	fmt.Println("logErrorsToStderr:", *logErrorsToStderr)
	fmt.Println("minioEndpoint:", *minioEndpoint)
	fmt.Println("minioAccessID:", *minioAccessID)
	fmt.Println("minioAccessSecret:", *minioAccessSecret)
	fmt.Println("minioSSL:", *minioSSL)
	fmt.Println("minioBucket:", *minioBucket)
	fmt.Println("minioLocation:", *minioLocation)
	fmt.Println("tail:", flag.Args())

	if *printVersion {
		fmt.Fprintf(os.Stderr, version+"\n")
		os.Exit(0)
	}

	if *printUsage {
		fmt.Fprintf(os.Stderr, "Usage of hapttic:\n")
		flag.PrintDefaults()
		os.Exit(0)
	}

	if utf8.RuneCountInString(*userScriptFileName) == 0 {
		log.Fatal("The path to the request handling script can not be empty.")
	}

	scriptFileName, err := filepath.Abs(*userScriptFileName)
	if err != nil {
		log.Fatal(err)
	}
	//Check if Minio should be used
	var useMinio = false
	useMinio = ensureMinioFlagsExist(*minioEndpoint, *minioAccessID, *minioAccessSecret)
	//Fill in Minio Details to Be passed to Hanlder
	minioFlags := minioDetails{
		minioEndpoint:     *minioEndpoint,
		minioAccessID:     *minioAccessID,
		minioAccessSecret: *minioAccessSecret,
		minioLocation:     *minioLocation,
		minioBucket:       *minioBucket,
		minioSSL:          *minioSSL,
		useMinio:          useMinio,
	}

	ensureRequestHandlingScriptExists(scriptFileName)
	fmt.Println("Use Minio?:", minioFlags.useMinio)

	http.HandleFunc("/", handleFuncWithScriptFileName(scriptFileName, *logErrorsToStderr, minioFlags))

	addr := *host + ":" + *port
	log.Println("Thanks for using hapttic v" + version)
	log.Println(fmt.Sprintf("Listening on %s", addr))
	log.Println(fmt.Sprintf("Forwarding requests to %s", scriptFileName))
	if *logErrorsToStderr {
		log.Println("Logging errors to stderr")
	}
	log.Fatal(http.ListenAndServe(addr, nil))
}
