package osm

import (
	"io"
	"net/http"
	"os"
)

func (i *Importer) download() error {
	i.logger.Infof("downloading %s", i.config.OSMURL)
	resp, err := http.Get(i.config.OSMURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	f, err := os.Create(i.config.OSMFilename)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}
