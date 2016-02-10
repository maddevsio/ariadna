package updater

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func downloadFromUrl(url string, destination string) error {
	logger.Info("Downloading file from %s, to %s", url, destination)

	output, err := os.Create(destination)
	if err != nil {
		fmt.Errorf("Error while creating %s: %s", destination, err)
	}
	defer output.Close()
	client := &http.Client{}

	request, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return fmt.Errorf("Error while downloading %s: %s", url, err)
	}

	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.3")

	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("Error while downloading %s: %s", url, err)
	}
	defer response.Body.Close()

	n, err := io.Copy(output, response.Body)

	if err != nil {
		return fmt.Errorf("Error while downloading %s: %s", url, err)
	}
	logger.Info("%d bytes downloaded", n)
	return nil
}
