{
  local policyRule = $.rbac.v1beta1.policyRule,
  local crd = $.apiextensions.v1beta1.customResourceDefinition,
  local deployment = $.apps.v1beta1.deployment,
  local container = $.core.v1.container,
  local containerPort = $.core.v1.containerPort,
  local volumeMount = $.core.v1.volumeMount,
  local service = $.core.v1.service,
  local servicePort = service.mixin.spec.portsType,
  local podAntiAffinity = deployment.mixin.spec.template.spec.affinity.podAntiAffinity,
  local weightedPodAffinityTerm = podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecutionType,

  contour_namespace:
    $.core.v1.namespace.new($._config.namespace),

  contour_crd:
    {
      apiVersion: 'apiextensions.k8s.io/v1beta1',
      kind: 'CustomResourceDefinition',
      metadata: {
        labels: {
          component: 'ingressroute',
        },
        name: 'ingressroutes.contour.heptio.com',
      },
      spec: {
        group: 'contour.heptio.com',
        names: {
          kind: 'IngressRoute',
          plural: 'ingressroutes',
        },
        scope: 'Namespaced',
        version: 'v1beta1',
      },
    },

  contour_rbac:
    $.util.rbac('contour', [
      policyRule.new() +
      policyRule.withApiGroups(['']) +
      policyRule.withResources([
        'configmaps',
        'endpoints',
        'nodes',
        'pods',
        'secrets',
      ]) +
      policyRule.withVerbs(['list', 'watch']),


      policyRule.new() +
      policyRule.withApiGroups(['']) +
      policyRule.withResources([
        'nodes',
      ]) +
      policyRule.withVerbs(['get']),


      policyRule.new() +
      policyRule.withApiGroups(['']) +
      policyRule.withResources([
        'services',
      ]) +
      policyRule.withVerbs(['get', 'list', 'watch']),


      policyRule.new() +
      policyRule.withApiGroups(['extensions']) +
      policyRule.withResources([
        'ingresses',
      ]) +
      policyRule.withVerbs(['get', 'list', 'watch']),


      policyRule.new() +
      policyRule.withApiGroups(['contour.heptio.com']) +
      policyRule.withResources([
        'ingressroutes',
      ]) +
      policyRule.withVerbs(['get', 'list', 'watch']),
    ]),


  contour_container_envoy::
    container.new('envoy', 'docker.io/envoyproxy/envoy-alpine:v1.6.0') +
    container.withPorts([
      $.core.v1.containerPort.new('http', 8080),
      $.core.v1.containerPort.new('https', 8443),
      $.core.v1.containerPort.new('http-metrics', 9001),
    ]) +
    container.withCommand('envoy') +
    container.withArgs([
      '-c',
      '/config/contour.yaml',
      '--service-cluster',
      'cluster0',
      '--service-node',
      'node0',
      '-l',
      'info',
      '--v2-config-only',
    ]),

  contour_container_contour::
    container.new('contour', 'gcr.io/heptio-images/contour:master') +
    container.withImagePullPolicy('Always') +
    container.withCommand('contour') +
    container.withArgs([
      'serve',
      '--incluster',
    ]),

  contour_countainer_init::
    container.new('envoy-initconfig', 'gcr.io/heptio-images/contour:master') +
    container.withImagePullPolicy('Always') +
    container.withCommand('contour') +
    container.withArgs([
      'bootstrap',
      '/config/contour.yaml',
    ]) +
    container.withVolumeMounts([volumeMount.new('contour-config', '/config')]),

  contour_deployment:
    deployment.new('contour', 2, [
      $.contour_container_envoy,
      $.contour_container_contour,
    ]) +
    deployment.mixin.spec.template.spec.withInitContainers([$.contour_countainer_init]) +
    $.util.emptyVolumeMount('contour-config', '/config') +
    deployment.mixin.spec.template.spec.withDnsPolicy('ClusterFirst') +
    deployment.mixin.spec.template.spec.withServiceAccount('contour') +
    deployment.mixin.spec.template.spec.withTerminationGracePeriodSeconds(30) +
    podAntiAffinity.withPreferredDuringSchedulingIgnoredDuringExecution(
      weightedPodAffinityTerm.withWeight(100) +
      weightedPodAffinityTerm.mixin.podAffinityTerm.withTopologyKey('kubernetes.io/hostname') +
      weightedPodAffinityTerm.mixin.podAffinityTerm.labelSelector.withMatchLabels(
        $.contour_deployment.spec.template.metadata.labels,
      ),
    ),


  contour_service:
    service.new(
      'contour',
      $.contour_deployment.spec.template.metadata.labels,
      [
        servicePort.newNamed('http', 80, 8080),
        servicePort.newNamed('https', 443, 8443),
      ],
    ) +
    service.mixin.metadata.withNamespace($._config.namespace),
}
