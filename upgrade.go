package usrbin

// CanSupportUpgrade
func (s SDK) CanSupportUpgrade() (bool, error) {
	for _, epm := range s.externalPackageManagers {
		isInstalled, err := epm.IsInstalled()
		if err != nil {
			return false, err
		}
		if isInstalled {
			return false, nil
		}
	}

	return true, nil
}

func (s SDK) ExternalUpgradeCommand() string {
	for _, epm := range s.externalPackageManagers {
		isInstalled, err := epm.IsInstalled()
		if err != nil {
			return ""
		}
		if isInstalled {
			return epm.UpgradeCommand()
		}
	}

	return ""
}
