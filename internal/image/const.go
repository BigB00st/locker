package image

const (
	imagesDir       = "/var/lib/locker/"
	imagesJsonFile  = imagesDir + "images.json"
	configFile      = "config.json"
	work            = "work"
	upper           = "upper"
	registry        = "https://registry-1.docker.io/v2/"
	authUrlIndex    = 1
	authHeaderIndex = 3
	idPrintLen      = 10
	lsPrintPad      = 23
	// Merged directory, mountpoint for container
	Merged = "merged"
)
