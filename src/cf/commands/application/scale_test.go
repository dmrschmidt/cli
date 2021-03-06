package application_test

import (
	"cf"
	"cf/api"
	. "cf/commands/application"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestScaleRequirements(t *testing.T) {
	args := []string{"-m", "1G", "my-app"}
	reqFactory, restarter, appRepo := getScaleDependencies()

	reqFactory.LoginSuccess = false
	reqFactory.TargetedSpaceSuccess = true
	callScale(t, args, reqFactory, restarter, appRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory.LoginSuccess = true
	reqFactory.TargetedSpaceSuccess = false
	callScale(t, args, reqFactory, restarter, appRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory.LoginSuccess = true
	reqFactory.TargetedSpaceSuccess = true
	callScale(t, args, reqFactory, restarter, appRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)
	assert.Equal(t, reqFactory.ApplicationName, "my-app")
}

func TestScaleFailsWithUsage(t *testing.T) {
	reqFactory, restarter, appRepo := getScaleDependencies()

	ui := callScale(t, []string{}, reqFactory, restarter, appRepo)

	assert.True(t, ui.FailedWithUsage)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestScaleAll(t *testing.T) {
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	reqFactory, restarter, appRepo := getScaleDependencies()
	reqFactory.Application = app

	ui := callScale(t, []string{"-i", "5", "-m", "512M", "my-app"}, reqFactory, restarter, appRepo)

	assert.Contains(t, ui.Outputs[0], "Scaling")
	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-space")
	assert.Contains(t, ui.Outputs[0], "my-user")

	assert.Equal(t, appRepo.ScaledApp.Guid, "my-app-guid")
	assert.Equal(t, appRepo.ScaledApp.Memory, uint64(512))
	assert.Equal(t, appRepo.ScaledApp.Instances, 5)
}

func TestScaleOnlyInstances(t *testing.T) {
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	reqFactory, restarter, appRepo := getScaleDependencies()
	reqFactory.Application = app

	callScale(t, []string{"-i", "5", "my-app"}, reqFactory, restarter, appRepo)

	assert.Equal(t, appRepo.ScaledApp.Guid, "my-app-guid")
	assert.Equal(t, appRepo.ScaledApp.DiskQuota, uint64(0))
	assert.Equal(t, appRepo.ScaledApp.Memory, uint64(0))
	assert.Equal(t, appRepo.ScaledApp.Instances, 5)
}

func TestScaleOnlyMemory(t *testing.T) {
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	reqFactory, restarter, appRepo := getScaleDependencies()
	reqFactory.Application = app

	callScale(t, []string{"-m", "512M", "my-app"}, reqFactory, restarter, appRepo)

	assert.Equal(t, appRepo.ScaledApp.Guid, "my-app-guid")
	assert.Equal(t, appRepo.ScaledApp.DiskQuota, uint64(0))
	assert.Equal(t, appRepo.ScaledApp.Memory, uint64(512))
	assert.Equal(t, appRepo.ScaledApp.Instances, 0)
}

func getScaleDependencies() (reqFactory *testreq.FakeReqFactory, restarter *testcmd.FakeAppRestarter, appRepo *testapi.FakeApplicationRepository) {
	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
	restarter = &testcmd.FakeAppRestarter{}
	appRepo = &testapi.FakeApplicationRepository{}
	return
}

func callScale(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, restarter *testcmd.FakeAppRestarter, appRepo api.ApplicationRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("scale", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	config := &configuration.Configuration{
		Space:        cf.Space{Name: "my-space"},
		Organization: cf.Organization{Name: "my-org"},
		AccessToken:  token,
	}

	cmd := NewScale(ui, config, restarter, appRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
