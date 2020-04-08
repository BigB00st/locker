package enviroment

// defaultFiles returns a list of files to copy from the host
// to the container.
func defaultFiles() []string {
	return []string{
		"/etc/resolv.conf",
	}
}
