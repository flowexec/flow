package updater

// Exported vars for whitebox access from the updater_test package.
var (
	GithubReleaseURL        = &githubReleaseURL
	GithubReleaseTagBaseURL = &githubReleaseTagBaseURL
	GithubDownloadBaseURL   = &githubDownloadBaseURL
	GetExecutablePath       = &getExecutablePath
	CurrentSemVer           = &currentSemVer
	CacheKey                = cacheKey
)

// Exported internal functions for whitebox testing.
var (
	AssetFileName    = assetFileName
	IsHomebrew       = isHomebrew
	ExtractFromTarGz = extractFromTarGz
	ExtractFromZip   = extractFromZip
	CopyReplace      = copyReplace
	UpgradeViaBinary = upgradeViaBinary
)
