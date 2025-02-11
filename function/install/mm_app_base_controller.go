package install

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/prokhorind/nextcloud/function/oauth"
)

//go:embed manifest.json
var manifestSource []byte

func GetManifest(c *gin.Context) {
	appType := os.Getenv("APP_TYPE")
	var manifest apps.Manifest
	json.Unmarshal(manifestSource, &manifest)

	if "HTTP" == appType {
		manifest.HTTP.RootURL = os.Getenv("APP_URL")
	}

	c.JSON(http.StatusOK, manifest)

}

func Ping(c *gin.Context) {

	c.JSON(http.StatusOK, apps.NewTextResponse("PONG"))
}

func Bindings(c *gin.Context) {
	creq := apps.CallRequest{}
	json.NewDecoder(c.Request.Body).Decode(&creq)

	commandBinding := apps.Binding{
		Icon:        "icon.png",
		Label:       "nextcloud",
		Description: "NextCloud App",
		Bindings:    []apps.Binding{},
	}

	token := oauth.Token{}
	remarshal(&token, creq.Context.OAuth2.User)

	if token.AccessToken == "" {
		commandBinding.Bindings = append(commandBinding.Bindings, apps.Binding{
			Location: "connect",
			Label:    "connect",
			Submit: apps.NewCall("/connect").WithExpand(apps.Expand{
				OAuth2App:             apps.ExpandAll,
				OAuth2User:            apps.ExpandAll,
				ActingUserAccessToken: apps.ExpandAll,
			}),
		})
	} else {
		commandBinding.Bindings = append(commandBinding.Bindings,
			apps.Binding{
				Location: "search",
				Label:    "search",
				Form: &apps.Form{
					Title: "Search Nextcloud files",
					Icon:  "icon.png",
					Fields: []apps.Field{
						{
							Type:       "text",
							Name:       "file_name",
							Label:      "file-name",
							IsRequired: true,
						},
					},
					Submit: apps.NewCall("/send").WithExpand(apps.Expand{
						OAuth2App:             apps.ExpandAll,
						OAuth2User:            apps.ExpandAll,
						Channel:               apps.ExpandAll,
						ActingUserAccessToken: apps.ExpandAll,
					}),
				},
			},
			apps.Binding{
				Location: "disconnect",
				Label:    "disconnect",
				Submit: apps.NewCall("/disconnect").WithExpand(apps.Expand{
					ActingUserAccessToken: apps.ExpandAll,
				}),
			},
			apps.Binding{
				Location: "create-calendar-event",
				Label:    "create-calendar-event",

				Submit: apps.NewCall("/create-calendar-event-form").WithExpand(apps.Expand{
					ActingUserAccessToken: apps.ExpandAll,
					OAuth2App:             apps.ExpandAll,
					OAuth2User:            apps.ExpandAll,
					Channel:               apps.ExpandAll,
				}),
			},

			apps.Binding{
				Location: "not-disturb",
				Label:    "not-disturb",
				Form: &apps.Form{
					Title: "Configures NextCloud client not disturb mode",
					Icon:  "icon.png",
					Fields: []apps.Field{
						{
							Type:       "bool",
							Name:       "enabled",
							Label:      "Enabled",
							IsRequired: true,
						},
					},
					Submit: apps.NewCall("/not-disturb").WithExpand(apps.Expand{
						ActingUserAccessToken: apps.ExpandAll,
						OAuth2App:             apps.ExpandAll,
						OAuth2User:            apps.ExpandAll,
						Channel:               apps.ExpandAll,
					}),
				},
			},

			apps.Binding{
				Location: "get-calendar-events",
				Label:    "get-calendar-events",

				Submit: apps.NewCall("/get-calendar-events-form").WithExpand(apps.Expand{
					ActingUserAccessToken: apps.ExpandAll,
					OAuth2App:             apps.ExpandAll,
					OAuth2User:            apps.ExpandAll,
					Channel:               apps.ExpandAll,
				}),
			},
		)
	}

	if creq.Context.ActingUser.IsSystemAdmin() {
		configure := apps.Binding{
			Location: "configure",
			Label:    "configure",
			Form: &apps.Form{
				Title: "Configures NextCloud client",
				Icon:  "icon.png",
				Fields: []apps.Field{
					{
						Type:       "text",
						Name:       "client_id",
						Label:      "client-id",
						IsRequired: true,
					},
					{
						Type:       "text",
						Name:       "client_secret",
						Label:      "client-secret",
						IsRequired: true,
					},

					{
						Type:       "text",
						Name:       "instance_url",
						Label:      "instance-url",
						IsRequired: true,
					},
				},
				Submit: apps.NewCall("/configure").WithExpand(apps.Expand{
					ActingUserAccessToken: apps.ExpandAll,
				}),
			},
		}
		commandBinding.Bindings = append(commandBinding.Bindings, configure)
	}

	upload := apps.Binding{
		Label:    "Upload file to Nextcloud",
		Location: apps.Location("id"),
		Icon:     "icon.png",
		Submit: apps.NewCall("/file-upload-form").WithExpand(apps.Expand{
			ActingUserAccessToken: apps.ExpandAll,
			OAuth2App:             apps.ExpandAll,
			OAuth2User:            apps.ExpandAll,
			Post:                  apps.ExpandAll,
		}),
	}

	c.JSON(http.StatusOK, apps.CallResponse{
		Type: apps.CallResponseTypeOK,
		Data: []apps.Binding{
			{
				Location: apps.LocationCommand,
				Bindings: []apps.Binding{
					commandBinding,
				},
			},
			{
				Location: apps.LocationPostMenu,
				Label:    "Nextcloud",
				Bindings: []apps.Binding{
					upload,
				},
			},
		}})
}

func remarshal(dst, src interface{}) {
	data, _ := json.Marshal(src)
	json.Unmarshal(data, dst)
}
