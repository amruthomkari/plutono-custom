package api

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/credativ/plutono/pkg/bus"
	"github.com/credativ/plutono/pkg/models"
	"github.com/credativ/plutono/pkg/services/licensing"
	"github.com/credativ/plutono/pkg/setting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setUpGetTeamMembersHandler() {
	bus.AddHandler("test", func(query *models.GetTeamMembersQuery) error {
		query.Result = []*models.TeamMemberDTO{
			{Email: "testUser@example.com", Login: testUserLogin},
			{Email: "user1@example.com", Login: "user1"},
			{Email: "user2@example.com", Login: "user2"},
		}
		return nil
	})
}

func TestTeamMembersAPIEndpoint_userLoggedIn(t *testing.T) {
	settings := setting.NewCfg()
	hs := &HTTPServer{
		Cfg:     settings,
		License: &licensing.OSSLicensingService{},
	}

	loggedInUserScenarioWithRole(t, "When calling GET on", "GET", "api/teams/1/members",
		"api/teams/:teamId/members", models.ROLE_ADMIN, func(sc *scenarioContext) {
			setUpGetTeamMembersHandler()

			sc.handlerFunc = hs.GetTeamMembers
			sc.fakeReqWithParams("GET", sc.url, map[string]string{}).exec()

			require.Equal(t, http.StatusOK, sc.resp.Code)

			var resp []models.TeamMemberDTO
			err := json.Unmarshal(sc.resp.Body.Bytes(), &resp)
			require.NoError(t, err)
			assert.Len(t, resp, 3)
		})

	t.Run("Given there is two hidden users", func(t *testing.T) {
		settings.HiddenUsers = map[string]struct{}{
			"user1":       {},
			testUserLogin: {},
		}
		t.Cleanup(func() { settings.HiddenUsers = make(map[string]struct{}) })

		loggedInUserScenarioWithRole(t, "When calling GET on", "GET", "api/teams/1/members",
			"api/teams/:teamId/members", models.ROLE_ADMIN, func(sc *scenarioContext) {
				setUpGetTeamMembersHandler()

				sc.handlerFunc = hs.GetTeamMembers
				sc.fakeReqWithParams("GET", sc.url, map[string]string{}).exec()

				require.Equal(t, http.StatusOK, sc.resp.Code)

				var resp []models.TeamMemberDTO
				err := json.Unmarshal(sc.resp.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.Len(t, resp, 2)
				assert.Equal(t, testUserLogin, resp[0].Login)
				assert.Equal(t, "user2", resp[1].Login)
			})
	})
}
