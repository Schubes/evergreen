package git_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/evergreen-ci/evergreen/agent/comm"
	agentutil "github.com/evergreen-ci/evergreen/agent/testutil"
	"github.com/evergreen-ci/evergreen/db"
	modelutil "github.com/evergreen-ci/evergreen/model/testutil"
	"github.com/evergreen-ci/evergreen/plugin"
	. "github.com/evergreen-ci/evergreen/plugin/builtin/git"
	"github.com/evergreen-ci/evergreen/plugin/plugintest"
	"github.com/evergreen-ci/evergreen/service"
	"github.com/evergreen-ci/evergreen/testutil"
	"github.com/mongodb/grip/slogger"
	. "github.com/smartystreets/goconvey/convey"
)

func init() {
	db.SetGlobalSessionProvider(db.SessionFactoryFromConfig(testutil.TestConfig()))
}

func TestGitPlugin(t *testing.T) {
	testConfig := testutil.TestConfig()
	Convey("With git plugin installed into plugin registry", t, func() {
		registry := plugin.NewSimpleRegistry()
		gitPlugin := &GitPlugin{}
		err := registry.Register(gitPlugin)
		testutil.HandleTestingErr(err, t, "Couldn't register plugin: %v")

		server, err := service.CreateTestServer(testConfig, nil, plugin.APIPlugins)
		testutil.HandleTestingErr(err, t, "Couldn't set up testing server")
		defer server.Close()

		configPath := filepath.Join(testutil.GetDirectoryOfFile(), "testdata", "plugin_clone.yml")
		patchPath := filepath.Join(testutil.GetDirectoryOfFile(), "testdata", "test.patch")

		modelData, err := modelutil.SetupAPITestData(testConfig, "test", "rhel55", configPath, modelutil.NoPatch)
		testutil.HandleTestingErr(err, t, "failed to setup test data")
		err = plugintest.SetupPatchData(modelData, patchPath, t)
		testutil.HandleTestingErr(err, t, "failed to setup patch data")
		taskConfig := modelData.TaskConfig

		httpCom := plugintest.TestAgentCommunicator(modelData, server.URL)

		logger := agentutil.NewTestLogger(slogger.StdOutAppender())
		Convey("all commands in test project should execute successfully", func() {
			for _, task := range taskConfig.Project.Tasks {
				So(len(task.Commands), ShouldNotEqual, 0)
				for _, command := range task.Commands {
					pluginCmds, err := registry.GetCommands(command, taskConfig.Project.Functions)
					testutil.HandleTestingErr(err, t, "Couldn't get plugin command: %v")
					So(pluginCmds, ShouldNotBeNil)
					So(err, ShouldBeNil)
					pluginCom := &comm.TaskJSONCommunicator{pluginCmds[0].Plugin(), httpCom}
					err = pluginCmds[0].Execute(logger, pluginCom, taskConfig, make(chan bool))
					So(err, ShouldBeNil)
				}
			}
			err = os.RemoveAll(taskConfig.WorkDir)
			testutil.HandleTestingErr(err, t, "Couldn't clean up test temp dir")
		})
	})
}
