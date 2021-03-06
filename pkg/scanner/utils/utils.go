package utils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Masterminds/semver/v3"

	"github.com/aquasecurity/fanal/types"

	"github.com/aquasecurity/trivy/pkg/log"
)

var (
	replacer = strings.NewReplacer(".alpha", "-alpha", ".beta", "-beta", ".rc", "-rc")
)

func MatchVersions(currentVersion *semver.Version, rangeVersions []string) bool {
	for _, v := range rangeVersions {
		v = replacer.Replace(v)
		constraintParts := strings.Split(v, ",")
		for j := range constraintParts {
			constraintParts[j] = FormatPatchVersion(constraintParts[j])
		}
		v = strings.Join(constraintParts, ",")
		c, err := semver.NewConstraint(v)
		if err != nil {
			log.Logger.Error("NewConstraint", "error", err)
			continue
		}
		// Validate a version against a constraint.
		valid, msgs := c.Validate(currentVersion)
		if valid {
			return true
		}
		for _, m := range msgs {
			// re-validate after removing the patch version
			if strings.HasSuffix(m.Error(), "is a prerelease version and the constraint is only looking for release versions") {
				v2, err := semver.NewVersion(fmt.Sprintf("%v.%v.%v", currentVersion.Major(), currentVersion.Minor(), currentVersion.Patch()))
				if err == nil {
					valid, _ = c.Validate(v2)
					if valid {
						return true
					}
				}
			}
		}
	}
	return false
}

func FormatVersion(pkg types.Package) string {
	return formatVersion(pkg.Epoch, pkg.Version, pkg.Release)
}

func FormatSrcVersion(pkg types.Package) string {
	return formatVersion(pkg.SrcEpoch, pkg.SrcVersion, pkg.SrcRelease)
}

func FormatPatchVersion(version string) string {
	part := strings.Split(version, ".")
	if len(part) > 3 {
		if _, err := strconv.Atoi(part[2]); err == nil {
			version = strings.Join(part[:3], ".") + "-" + strings.Join(part[3:], ".")
		}
	}
	return version
}

func formatVersion(epoch int, version, release string) string {
	v := version
	if release != "" {
		v = fmt.Sprintf("%s-%s", v, release)
	}
	if epoch != 0 {
		v = fmt.Sprintf("%d:%s", epoch, v)
	}
	return v

}
