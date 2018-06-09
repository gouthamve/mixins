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

  certmanager_namespace:
    $.core.v1.namespace.new($._config.namespace),

  certificate_crd:
    {
      apiVersion: 'apiextensions.k8s.io/v1beta1',
      kind: 'CustomResourceDefinition',
      metadata: {
        labels: {
          app: 'cert-manager',
        },
        name: 'certificates.certmanager.k8s.io',
      },
      spec: {
        group: 'certmanager.k8s.io',
        names: {
          kind: 'Certificate',
          plural: 'certificates',
          shortNames: ['cert', 'certs'],
        },
        scope: 'Namespaced',
        version: 'v1alpha1',
      },
    },

  clusterissuer_crd:
    {
      apiVersion: 'apiextensions.k8s.io/v1beta1',
      kind: 'CustomResourceDefinition',
      metadata: {
        labels: {
          app: 'cert-manager',
        },
        name: 'clusterissuers.certmanager.k8s.io',
      },
      spec: {
        group: 'certmanager.k8s.io',
        names: {
          kind: 'ClusterIssuer',
          plural: 'clusterissuers',
        },
        scope: 'Cluster',
        version: 'v1alpha1',
      },
    },

  issuer_crd:
    {
      apiVersion: 'apiextensions.k8s.io/v1beta1',
      kind: 'CustomResourceDefinition',
      metadata: {
        labels: {
          app: 'cert-manager',
        },
        name: 'issuers.certmanager.k8s.io',
      },
      spec: {
        group: 'certmanager.k8s.io',
        names: {
          kind: 'Issuer',
          plural: 'issuers',
        },
        scope: 'Namespaced',
        version: 'v1alpha1',
      },
    },

  certmanager_rbac:
    $.util.rbac('cert-manager', [
      policyRule.new() +
      policyRule.withApiGroups(['certmanager.k8s.io']) +
      policyRule.withResources([
        'certificates',
        'issuers',
        'clusterissuers',
      ]) +
      policyRule.withVerbs(['*']),


      policyRule.new() +
      policyRule.withApiGroups(['']) +
      policyRule.withResources([
        'configmaps',
        'secrets',
        'events',
        'services',
        'pods',
      ]) +
      policyRule.withVerbs(['*']),


      policyRule.new() +
      policyRule.withApiGroups(['extensions']) +
      policyRule.withResources([
        'ingresses',
      ]) +
      policyRule.withVerbs(['*']),
    ]),


  certmanager_container::
    container.new('cert-manager', 'quay.io/jetstack/cert-manager-controller:v0.3.0') +
    container.withPorts([
      $.core.v1.containerPort.new('http-metrics', 9402),
    ]) +
    container.withArgs([
      '--cluster-resource-namespace=' + $._config.namespace,
      '--leader-election-namespace=' + $._config.namespace,
    ]) +
    $.util.resourcesRequests('10m', '32Mi'),

  certmanager_deployment:
    deployment.new('cert-manager', 1, [
      $.certmanager_container,
    ]) +
    deployment.mixin.spec.template.spec.withServiceAccount('cert-manager'),
}
