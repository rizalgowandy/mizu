package app

import (
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/up9inc/mizu/agent/pkg/api"
	"github.com/up9inc/mizu/agent/pkg/config"
	"github.com/up9inc/mizu/agent/pkg/middlewares"
	"github.com/up9inc/mizu/agent/pkg/routes"
	"github.com/up9inc/mizu/shared/logger"
	tapApi "github.com/up9inc/mizu/tap/api"
)

var (
	ConfigRoutes     *gin.RouterGroup
	UserRoutes       *gin.RouterGroup
	InstallRoutes    *gin.RouterGroup
	OASRoutes        *gin.RouterGroup
	ServiceMapRoutes *gin.RouterGroup
	QueryRoutes      *gin.RouterGroup
	EntriesRoutes    *gin.RouterGroup
	MetadataRoutes   *gin.RouterGroup
	StatusRoutes     *gin.RouterGroup
)

func HostApi(socketHarOutputChannel chan<- *tapApi.OutputChannelItem) *gin.Engine {
	app := gin.Default()

	app.GET("/echo", func(c *gin.Context) {
		c.String(http.StatusOK, "Here is Mizu agent")
	})

	eventHandlers := api.RoutesEventHandlers{
		SocketOutChannel: socketHarOutputChannel,
	}

	app.Use(disableRootStaticCache())

	var staticFolder string
	if config.Config.StandaloneMode {
		staticFolder = "./site-standalone"
	} else {
		staticFolder = "./site"
	}

	indexStaticFile := staticFolder + "/index.html"
	if err := setUIFlags(indexStaticFile); err != nil {
		logger.Log.Errorf("Error setting ui flags, err: %v", err)
	}

	app.Use(static.ServeRoot("/", staticFolder))
	app.NoRoute(func(c *gin.Context) {
		c.File(indexStaticFile)
	})

	app.Use(middlewares.CORSMiddleware()) // This has to be called after the static middleware, does not work if its called before

	api.WebSocketRoutes(app, &eventHandlers, startTime)

	if config.Config.StandaloneMode {
		ConfigRoutes = routes.ConfigRoutes(app)
		UserRoutes = routes.UserRoutes(app)
		InstallRoutes = routes.InstallRoutes(app)
	}
	if config.Config.OAS {
		OASRoutes = routes.OASRoutes(app)
	}
	if config.Config.ServiceMap {
		ServiceMapRoutes = routes.ServiceMapRoutes(app)
	}

	QueryRoutes = routes.QueryRoutes(app)
	EntriesRoutes = routes.EntriesRoutes(app)
	MetadataRoutes = routes.MetadataRoutes(app)
	StatusRoutes = routes.StatusRoutes(app)

	return app
}

func disableRootStaticCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.RequestURI == "/" {
			// Disable cache only for the main static route
			c.Writer.Header().Set("Cache-Control", "no-store")
		}

		c.Next()
	}
}

func setUIFlags(uiIndexPath string) error {
	read, err := ioutil.ReadFile(uiIndexPath)
	if err != nil {
		return err
	}

	replacedContent := strings.Replace(string(read), "__IS_OAS_ENABLED__", strconv.FormatBool(config.Config.OAS), 1)
	replacedContent = strings.Replace(replacedContent, "__IS_SERVICE_MAP_ENABLED__", strconv.FormatBool(config.Config.ServiceMap), 1)

	err = ioutil.WriteFile(uiIndexPath, []byte(replacedContent), 0)
	if err != nil {
		return err
	}

	return nil
}
