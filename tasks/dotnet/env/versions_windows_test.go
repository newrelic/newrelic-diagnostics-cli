package env

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Dotnet/Env/Versions", func() {
	Describe("releaseKeyToVersion()", func() {
		DescribeTable("Should map the Release registry value to a .NET Framework version",
			func(release uint64, expected string) {
				Expect(releaseKeyToVersion(release)).To(Equal(expected))
			},
			Entry("below 4.5", uint64(378388), ""),
			Entry("4.5", uint64(378389), "4.5"),
			Entry("4.5.1", uint64(378675), "4.5.1"),
			Entry("4.5.1 on Windows 8.1", uint64(378758), "4.5.1"),
			Entry("4.5.2", uint64(379893), "4.5.2"),
			Entry("4.6", uint64(393295), "4.6"),
			Entry("4.6.1", uint64(394254), "4.6.1"),
			Entry("4.6.2", uint64(394802), "4.6.2"),
			Entry("4.7", uint64(460798), "4.7"),
			Entry("4.7.1", uint64(461308), "4.7.1"),
			Entry("4.7.2", uint64(461808), "4.7.2"),
			Entry("4.8", uint64(528040), "4.8"),
			Entry("4.8 on other OSes", uint64(528372), "4.8"),
			Entry("4.8.1 on Windows 11 22H2", uint64(533320), "4.8.1"),
			Entry("4.8.1 on other OSes", uint64(533325), "4.8.1"),
			Entry("newer than 4.8.1", uint64(533326), "4.8.1 or later"),
		)
	})
})
