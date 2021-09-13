package helm_installer

func (i *HelmInstaller) Exit(err error) {
	if err != nil {
		i.Log.Fatalf("Failure: %v", err)
	}
}
