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
	pluginMQTTWorker      = "MQTTWorker"
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
			stompbroker.NewSTOMPBroker(stompbroker.InitDefaultConnection("127.0.0.1:61613", "[username]", "[password]")),`,
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

		pluginMQTTWorker: {
			name:        pluginMQTTWorker,
			packageName: "github.com/golangid/candi-plugin/mqtt-broker",
			editConfig: map[string]string{
				`import (`: `import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	mqttbroker "github.com/golangid/candi-plugin/mqtt-broker"`,
				"brokerDeps := broker.InitBrokers(": `brokerDeps := broker.InitBrokers(
			mqttbroker.NewMQTTBroker(mqtt.NewClientOptions().
				AddBroker("tcp://127.0.0.1:1883").
				SetClientID("MQTTClientID").
				SetUsername("MQTTUsername").
				SetPassword("MQTTPassword").
				SetCleanSession(false).
				SetAutoReconnect(true).
				SetConnectRetry(true),
			),`,
			},
			editAppFactory: map[string]string{
				`import (`: `import (
	mqttbroker "github.com/golangid/candi-plugin/mqtt-broker"`,
				`return
}`: `apps = append(apps, mqttbroker.NewMQTTSubscriber(
		service,
		service.GetDependency().GetBroker(mqttbroker.MQTTBroker),
	))
	return
}`,
			},
			editModule: map[string]string{
				"import (": `import (
	mqttbroker "github.com/golangid/candi-plugin/mqtt-broker"
`,
				"mod.workerHandlers = map[types.Worker]interfaces.WorkerHandler{": `mod.workerHandlers = map[types.Worker]interfaces.WorkerHandler{
		mqttbroker.MQTTBroker: workerhandler.NewMQTTWorkerHandler(usecase.GetSharedUsecase(), deps),`,
			},
		},
	}
)
