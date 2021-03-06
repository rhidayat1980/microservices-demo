(import "ksonnet-util/kausal.libsonnet") +
{

  local deploy = $.apps.v1.deployment,
  local container = $.core.v1.container,
  local port = $.core.v1.containerPort,
  local service = $.core.v1.service,

  _config+:: {
    shippingservice: {
      app: "shippingservice",
      namespace: $._config.namespace, //set a default namespace if not overrided in the main file
      port: 50051,
      portName: "grpc",
      ports: [{ portName: "health", port: 50053 }],
      image: {
        repo: $._config.image.repo,
        name: "shippingservice",
        tag: $._config.image.tag,
      },
      labels: {app: "shippingservice"},
      env: {
        PORT: "%s" % $._config.shippingservice.port,
        HEALTH_PORT: "%s" % $._config.shippingservice.ports[0].port,
    },
      readinessProbe: container.mixin.readinessProbe.exec.withCommand(["/bin/grpc_health_probe", "-addr=:%s" % self.ports[0].port,]),
      livenessProbe: container.mixin.livenessProbe.exec.withCommand(["/bin/grpc_health_probe", "-addr=:%s" % self.ports[0].port,]),
      limits: container.mixin.resources.withLimits({cpu: "200m", memory: "128Mi"}),
      requests: container.mixin.resources.withRequests({cpu: "100m", memory: "64Mi"}),
      deploymentExtra: {},
      serviceExtra: {},
    },
  },
}