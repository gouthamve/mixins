{
  letsencrypt_staging:
    {
      apiVersion: 'certmanager.k8s.io/v1alpha1',
      kind: 'ClusterIssuer',
      metadata: {
        name: 'letsencrypt-staging',
        namespace: 'cert-manager',
      },
      spec: {
        acme: {
          email: $._config.letsencrypt_email,
          http01: {},
          privateKeySecretRef: {
            name: 'letsencrypt-staging',
          },
          server: 'https://acme-staging-v02.api.letsencrypt.org/directory',
        },
      },
    },


  letsencrypt_prod:
    {
      apiVersion: 'certmanager.k8s.io/v1alpha1',
      kind: 'ClusterIssuer',
      metadata: {
        name: 'letsencrypt-prod',
        namespace: 'cert-manager',
      },
      spec: {
        acme: {
          email: $._config.letsencrypt_email,
          http01: {},
          privateKeySecretRef: {
            name: 'letsencrypt-prod',
          },
          server: 'https://acme-v02.api.letsencrypt.org/directory',
        },
      },
    },
}
