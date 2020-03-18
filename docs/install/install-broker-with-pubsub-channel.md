# Installing Broker with Pub/Sub Channel

In Knative Eventing, a Broker represents an "event mesh". Events are sent to the
Broker's ingress and are then sent to any subscribers that are interested in
that event. Brokers are currently backed by a configurable Channel. In this
guide, we show an example of how to configure a Broker backed by the Pub/Sub
Channel.

## Prerequisites

1. [Install Knative-GCP](./install-knative-gcp.md). Remember to install
   [Knative Eventing](https://knative.dev/docs/eventing/).

1. [Create a Pub/Sub enabled Service Account](./pubsub-service-account.md).

## Deployment

1.  Verity the configmap `config-br-default-channel` is in the `cloud-run-events` namespace.
    ```shell
    kubectl get configmap config-br-default-channel -n cloud-run-events
    ```
    By default, it is assumed to use default secret. Modify [`config-br-default-channel`](config-br-default-channel.yaml) 
    to use different method and apply it.
       1. If you are using workload identity, update `serviceAccount` with the Pub/Sub enabled service account you created in [Create a Pub/Sub enabled Service Account](./pubsub-service-account.md).
        
       1. If you are using non-default secret, update `project` and `secret`.
       ```shell
       kubectl apply -f config-br-default-channel.yaml
       ```
    
1.  Patch the configmap in the `knative-eventing` namespace to use the Pub/Sub
    `Channel` as the
    [default channel](https://knative.dev/docs/eventing/channel-based-broker/)
    for Brokers with
    [patch-config-br-defaults-with-pubsub.yaml](./patch-config-br-defaults-with-pubsub.yaml).

    ```shell
    kubectl patch configmap config-br-defaults -n knative-eventing --patch "$(cat patch-config-br-defaults-with-pubsub.yaml)"
    ```

1.  Add the `knative-eventing-injection` label to your namespace with the
    following command.

    ```shell
    kubectl label namespace default knative-eventing-injection=enabled
    ```

    This triggers a reconciliation process that creates the `default Broker` in
    that namespace.

1.  Verify that the `Broker` is running

    ```shell
    kubectl --namespace default get broker default
    ```

    This shows the `Broker` that you created:

    ```shell
    NAME      READY   REASON   URL                                                        AGE
    default   True             http://default-broker.default.svc.cluster.local      1m
    ```

    When the `Broker` has the `READY=True` state, it can start processing any
    events it receives.
