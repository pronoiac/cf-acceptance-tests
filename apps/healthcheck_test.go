package apps

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	"github.com/cloudfoundry/cf-acceptance-tests/helpers/app_helpers"
	"github.com/cloudfoundry/cf-acceptance-tests/helpers/assets"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
)

var _ = Describe(deaUnsupportedTag+"Healthcheck", func() {
	var appName string

	BeforeEach(func() {
		appName = generator.PrefixedRandomName("CATS-APP-")
	})

	AfterEach(func() {
		app_helpers.AppReport(appName, DEFAULT_TIMEOUT)
		Eventually(cf.Cf("delete", appName, "-f"), DEFAULT_TIMEOUT).Should(Exit(0))
	})

	Describe("when the healthcheck is set to none", func() {
		It("starts up successfully", func() {
			By("pushing it")
			Eventually(cf.Cf(
				"push", appName,
				"-p", assets.NewAssets().WorkerApp,
				"--no-start",
				"-b", "go_buildpack",
				"-m", DEFAULT_MEMORY_LIMIT,
				"-d", config.AppsDomain,
				"-i", "1",
				"-u", "none"),
				DEFAULT_TIMEOUT,
			).Should(Exit(0))

			By("staging and running it")
			app_helpers.SetBackend(appName)
			Eventually(cf.Cf("start", appName), CF_PUSH_TIMEOUT).Should(Exit(0))

			By("verifying it's up")
			Eventually(func() *Session {
				appLogsSession := cf.Cf("logs", "--recent", appName)
				Expect(appLogsSession.Wait(DEFAULT_TIMEOUT)).To(Exit(0))
				return appLogsSession
			}, DEFAULT_TIMEOUT).Should(gbytes.Say("I am working at"))
		})
	})

	Describe("when the healthcheck is set to port", func() {
		It("starts up successfully", func() {
			By("pushing it")
			Eventually(cf.Cf(
				"push", appName,
				"-p", assets.NewAssets().Dora,
				"--no-start",
				"-b", "ruby_buildpack",
				"-m", DEFAULT_MEMORY_LIMIT,
				"-d", config.AppsDomain,
				"-i", "1",
				"-u", "port"),
				DEFAULT_TIMEOUT,
			).Should(Exit(0))

			By("staging and running it")
			app_helpers.SetBackend(appName)
			Eventually(cf.Cf("start", appName), CF_PUSH_TIMEOUT).Should(Exit(0))

			By("verifying it's up")
			Eventually(helpers.CurlingAppRoot(appName)).Should(ContainSubstring("Hi, I'm Dora!"))
		})
	})
})
