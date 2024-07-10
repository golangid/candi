package main

type plugin struct {
	name           string
	packageName    string
	editConfig     map[string]string
	editAppFactory map[string]string
	editModule     map[string]string
}

const (
	pluginGCPPubSubWorker = "GCPPubSubWorker"
	pluginSTOMPWorker     = "STOMPWorker"
)

var (
	plugins = map[string]plugin{
		pluginGCPPubSubWorker: {
			name:        pluginGCPPubSubWorker,
			packageName: "github.com/golangid/candi-plugin/gcppubsub",
			editConfig: map[string]string{
				`import (`: `import (
	"github.com/golangid/candi-plugin/gcppubsub"`,
				"brokerDeps := broker.InitBrokers(": `brokerDeps := broker.InitBrokers(
			gcppubsub.NewGCPPubSubBroker(
				gcppubsub.BrokerSetClient(gcppubsub.InitDefaultClient("[gcp-project-id]", "credentials-file.json")),
			),`,
			},
			editAppFactory: map[string]string{
				`import (`: `import (
	"github.com/golangid/candi-plugin/gcppubsub"`,
				`return
}`: `apps = append(apps, gcppubsub.NewPubSubWorker(
		service,
		service.GetDependency().GetBroker(gcppubsub.GoogleCloudPubSub),
		"candi-example",
	))
	return
}`,
			},
			editModule: map[string]string{
				"import (": `import (
	"github.com/golangid/candi-plugin/gcppubsub"
`,
				"mod.workerHandlers = map[types.Worker]interfaces.WorkerHandler{": `mod.workerHandlers = map[types.Worker]interfaces.WorkerHandler{
		gcppubsub.GoogleCloudPubSub: workerhandler.NewGCPPubSubWorkerHandler(usecase.GetSharedUsecase(), deps),`,
			},
		},

		pluginSTOMPWorker: {
			name:        pluginSTOMPWorker,
			packageName: "github.com/golangid/candi-plugin/stomp-broker",
			editConfig: map[string]string{
				`import (`: `import (
	stompbroker "github.com/golangid/candi-plugin/stomp-broker"`,
				"brokerDeps := broker.InitBrokers(": `brokerDeps := broker.InitBrokers(
			stompbroker.NewSTOMPBroker(stompbroker.InitDefaultConnection("[broker host]", "[username]", "[password]")),`,
			},
			editAppFactory: map[string]string{
				`import (`: `import (
	stompbroker "github.com/golangid/candi-plugin/stomp-broker"`,
				`return
}`: `apps = append(apps, stompbroker.NewSTOMPWorker(
		service,
		service.GetDependency().GetBroker(stompbroker.STOMPBroker),
	))
	return
}`,
			},
			editModule: map[string]string{
				"import (": `import (
	stompbroker "github.com/golangid/candi-plugin/stomp-broker"
`,
				"mod.workerHandlers = map[types.Worker]interfaces.WorkerHandler{": `mod.workerHandlers = map[types.Worker]interfaces.WorkerHandler{
		stompbroker.STOMPBroker: workerhandler.NewSTOMPWorkerHandler(usecase.GetSharedUsecase(), deps),`,
			},
		},
	}
)
