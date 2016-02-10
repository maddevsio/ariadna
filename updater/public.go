package updater

func DownloadOSMFile(url string, destination string) error {
	err := downloadFromUrl(url, destination)
	if err != nil {
		return err
	}
	return nil
}
