package app

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/antelman107/net-wait-go/wait"
	"github.com/gin-gonic/gin"
	"github.com/op/go-logging"
	basenine "github.com/up9inc/basenine/client/go"
	"github.com/up9inc/mizu/agent/pkg/api"
	"github.com/up9inc/mizu/agent/pkg/config"
	"github.com/up9inc/mizu/agent/pkg/elastic"
	"github.com/up9inc/mizu/agent/pkg/oas"
	"github.com/up9inc/mizu/agent/pkg/servicemap"
	"github.com/up9inc/mizu/agent/pkg/up9"
	"github.com/up9inc/mizu/shared"
	"github.com/up9inc/mizu/shared/logger"
	tapApi "github.com/up9inc/mizu/tap/api"
)

var (
	Extensions    []*tapApi.Extension          // global
	ExtensionsMap map[string]*tapApi.Extension // global
	startTime     int64
)

func RunInApiServerMode(namespace string) *gin.Engine {
	configureBasenineServer(shared.BasenineHost, shared.BaseninePort)
	startTime = time.Now().UnixNano() / int64(time.Millisecond)
	api.StartResolving(namespace)

	outputItemsChannel := make(chan *tapApi.OutputChannelItem)
	filteredOutputItemsChannel := make(chan *tapApi.OutputChannelItem)
	enableExpFeatureIfNeeded()
	go FilterItems(outputItemsChannel, filteredOutputItemsChannel)
	go api.StartReadingEntries(filteredOutputItemsChannel, nil, ExtensionsMap)

	syncEntriesConfig := getSyncEntriesConfig()
	if syncEntriesConfig != nil {
		if err := up9.SyncEntries(syncEntriesConfig); err != nil {
			logger.Log.Error("Error syncing entries, err: %v", err)
		}
	}

	return HostApi(outputItemsChannel)
}

func configureBasenineServer(host string, port string) {
	if !wait.New(
		wait.WithProto("tcp"),
		wait.WithWait(200*time.Millisecond),
		wait.WithBreak(50*time.Millisecond),
		wait.WithDeadline(5*time.Second),
		wait.WithDebug(config.Config.LogLevel == logging.DEBUG),
	).Do([]string{fmt.Sprintf("%s:%s", host, port)}) {
		logger.Log.Panicf("Basenine is not available!")
	}

	// Limit the database size to default 200MB
	err := basenine.Limit(host, port, config.Config.MaxDBSizeBytes)
	if err != nil {
		logger.Log.Panicf("Error while limiting database size: %v", err)
	}

	// Define the macros
	for _, extension := range Extensions {
		macros := extension.Dissector.Macros()
		for macro, expanded := range macros {
			err = basenine.Macro(host, port, macro, expanded)
			if err != nil {
				logger.Log.Panicf("Error while adding a macro: %v", err)
			}
		}
	}
}

func getSyncEntriesConfig() *shared.SyncEntriesConfig {
	syncEntriesConfigJson := os.Getenv(shared.SyncEntriesConfigEnvVar)
	if syncEntriesConfigJson == "" {
		return nil
	}

	var syncEntriesConfig = &shared.SyncEntriesConfig{}
	err := json.Unmarshal([]byte(syncEntriesConfigJson), syncEntriesConfig)
	if err != nil {
		panic(fmt.Sprintf("env var %s's value of %s is invalid! json must match the shared.SyncEntriesConfig struct, err: %v", shared.SyncEntriesConfigEnvVar, syncEntriesConfigJson, err))
	}

	return syncEntriesConfig
}

func FilterItems(inChannel <-chan *tapApi.OutputChannelItem, outChannel chan *tapApi.OutputChannelItem) {
	for message := range inChannel {
		if message.ConnectionInfo.IsOutgoing && api.CheckIsServiceIP(message.ConnectionInfo.ServerIP) {
			continue
		}

		outChannel <- message
	}
}

func enableExpFeatureIfNeeded() {
	if config.Config.OAS {
		oas.GetOasGeneratorInstance().Start()
	}
	if config.Config.ServiceMap {
		servicemap.GetInstance().SetConfig(config.Config)
	}
	elastic.GetInstance().Configure(config.Config.Elastic)
}
