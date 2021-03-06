package v3

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	"github.com/cloudfoundry/cf-acceptance-tests/helpers/assets"
	. "github.com/cloudfoundry/cf-acceptance-tests/helpers/v3_helpers"
)

type ProcessStats struct {
	Instance []struct {
		State string `json:"state"`
	} `json:"resources"`
}

var _ = Describe("process", func() {
	var (
		appName     string
		appGuid     string
		packageGuid string
		spaceGuid   string
		token       string
	)

	BeforeEach(func() {
		appName = generator.PrefixedRandomName("CATS-APP-")
		spaceGuid = GetSpaceGuidFromName(context.RegularUserContext().Space)
		appGuid = CreateApp(appName, spaceGuid, `{"foo":"bar"}`)
		packageGuid = CreatePackage(appGuid)
		token := GetAuthToken()
		uploadUrl := fmt.Sprintf("%s%s/v3/packages/%s/upload", config.Protocol(), config.ApiEndpoint, packageGuid)
		UploadPackage(uploadUrl, assets.NewAssets().DoraZip, token)
		WaitForPackageToBeReady(packageGuid)
	})

	AfterEach(func() {
		FetchRecentLogs(appGuid, token, config)
		DeleteApp(appGuid)
	})

	Describe("terminating an instance", func() {
		var (
			index       = 0
			processType = "web"
			webProcess  Process
		)

		BeforeEach(func() {
			dropletGuid := StageBuildpackPackage(packageGuid, "ruby_buildpack")
			WaitForDropletToStage(dropletGuid)

			AssignDropletToApp(appGuid, dropletGuid)

			processes := GetProcesses(appGuid, appName)
			webProcess = GetProcessByType(processes, "web")

			CreateAndMapRoute(appGuid, context.RegularUserContext().Space, helpers.LoadConfig().AppsDomain, webProcess.Name)

			StartApp(appGuid)

			Eventually(func() string {
				return helpers.CurlAppRoot(webProcess.Name)
			}, DEFAULT_TIMEOUT).Should(ContainSubstring("Hi, I'm Dora!"))

			Expect(cf.Cf("apps").Wait(DEFAULT_TIMEOUT)).To(Say(fmt.Sprintf("%s\\s+started", webProcess.Name)))
		})

		Context("/v3/apps/:guid/processes/:type/instances/:index", func() {
			It("restarts the instance", func() {
				statsUrl := fmt.Sprintf("/v3/apps/%s/processes/web/stats", appGuid)

				By("ensuring the instance is running")
				statsBody := cf.Cf("curl", statsUrl).Wait(DEFAULT_TIMEOUT).Out.Contents()
				statsJSON := ProcessStats{}
				json.Unmarshal(statsBody, &statsJSON)
				Expect(statsJSON.Instance[0].State).To(Equal("RUNNING"))

				By("terminating the instance")
				terminateUrl := fmt.Sprintf("/v3/apps/%s/processes/%s/instances/%d", appGuid, processType, index)
				cf.Cf("curl", terminateUrl, "-X", "DELETE").Wait(DEFAULT_TIMEOUT)

				By("ensuring the instance is no longer running")
				// Note that this depends on a 30s run loop waking up in Diego.
				Eventually(func() string {
					statsBodyAfter := cf.Cf("curl", statsUrl).Wait(DEFAULT_TIMEOUT).Out.Contents()
					json.Unmarshal(statsBodyAfter, &statsJSON)
					return statsJSON.Instance[0].State
				}, 35*time.Second).ShouldNot(Equal("RUNNING"))

				By("ensuring the instance is running again")
				Eventually(func() string {
					statsBodyAfter := cf.Cf("curl", statsUrl).Wait(DEFAULT_TIMEOUT).Out.Contents()
					json.Unmarshal(statsBodyAfter, &statsJSON)
					return statsJSON.Instance[0].State
				}, 35*time.Second).Should(Equal("RUNNING"))
			})
		})

		Context("/v3/processes/:guid/instances/:index", func() {
			It("restarts the instance", func() {
				statsUrl := fmt.Sprintf("/v3/apps/%s/processes/web/stats", appGuid)

				By("ensuring the instance is running")
				statsBody := cf.Cf("curl", statsUrl).Wait(DEFAULT_TIMEOUT).Out.Contents()
				statsJSON := ProcessStats{}
				json.Unmarshal(statsBody, &statsJSON)
				Expect(statsJSON.Instance[0].State).To(Equal("RUNNING"))

				By("terminating the instance")
				terminateUrl := fmt.Sprintf("/v3/processes/%s/instances/%d", webProcess.Guid, index)
				cf.Cf("curl", terminateUrl, "-X", "DELETE").Wait(DEFAULT_TIMEOUT)

				By("ensuring the instance is no longer running")
				// Note that this depends on a 30s run loop waking up in Diego.
				Eventually(func() string {
					statsBodyAfter := cf.Cf("curl", statsUrl).Wait(DEFAULT_TIMEOUT).Out.Contents()
					json.Unmarshal(statsBodyAfter, &statsJSON)
					return statsJSON.Instance[0].State
				}, 35*time.Second).ShouldNot(Equal("RUNNING"))

				By("ensuring the instance is running again")
				Eventually(func() string {
					statsBodyAfter := cf.Cf("curl", statsUrl).Wait(DEFAULT_TIMEOUT).Out.Contents()
					json.Unmarshal(statsBodyAfter, &statsJSON)
					return statsJSON.Instance[0].State
				}, 35*time.Second).Should(Equal("RUNNING"))
			})
		})
	})
})
