# Cloud AuditLogsSource Example

## Overview

This sample shows how to Configure a `AuditLogsSource` resource to 
Reads data from [Cloud Audit Logs](https://cloud.google.com/logging/docs/audit/) and directly publishes to the underlying transport (Pub/Sub), 
in CloudEvents format.

## Prerequisites

1. [Install Knative with GCP](../install/README.md).

1. [Create a Pub/Sub enabled Service Account](../install/pubsub-service-account.md)

1. Enable the `Cloud Audit Logs API` on your project:

   ```shell
   gcloud services enable logging.googleapis.com 
   gcloud services enable stackdriver.googleapis.com 
   ```

## Deployment

1. Create a [auditlogssource](./auditlogssource.yaml). This `AuditLogsSource` will get Cloud Audit Log Entries for topic creation.  
   You can change the `serviceName`, `methodName` and `resourceName` to select the Cloud Audit Log Entries you want to view. 
           |   AuditLogsSource Spec  |       Audit Log Entry Fields   |
           |:----------------:	|:--------------------------:	|
           |      serviceName   |     protoPayload.serviceName  |
           |      methodName    |     protoPayload.methodName   |
           |     resourceName   |     protoPayload.resourceName |
   
   ```shell
      kubectl apply --filename auditlogssource.yaml
   ```    


1. Create a [service](./event-display.yaml) that the AuditLogsSource will sink into:

   ```shell
   kubectl apply --filename event-display.yaml
   ```
   
## Publish

Create a GCP PubSub Topic:

   ```shell
   gcloud pubsub topics create test-auditlogs-source
   ```


## Verify

We will verify that the published event was sent by looking at the logs of the
service that this AuditLogsSource sinks to.

1. We need to wait for the downstream pods to get started and receive our event,
   wait 60 seconds.

   - You can check the status of the downstream pods with:

     ```shell
     kubectl get pods --selector app=event-display
     ```

     You should see at least one.

1. Inspect the logs of the service:

   ```shell
   kubectl logs --selector app=event-display -c user-container --tail=200
   ```

You should see log lines similar to:

```shell
☁️  cloudevents.Event
Validation: valid
Context Attributes,
  specversion: 1.0
  type: com.google.cloud.auditlog.event
  source: pubsub.googleapis.com/projects/test
  subject: pubsub.googleapis.com/projects/test/topics/test-auditlogs-source
  id: 8c2iznd54odprojects/test/logs/cloudaudit.googleapis.com%2Factivity2020-01-07T20:56:30.516179081Z
  time: 2020-01-07T20:56:30.516179081Z
  dataschema: type.googleapis.com/google.logging.v2.LogEntry
  datacontenttype: application/json
Extensions,
  methodname: google.pubsub.v1.Publisher.CreateTopic
  resourcename: projects/test/topics/test-auditlogs-source
  servicename: pubsub.googleapis.com
Data,
  {
    "insertId": "8c2iznd54od",
    "logName": "projects/test/logs/cloudaudit.googleapis.com%2Factivity",
    "protoPayload": {
      "@type": "type.googleapis.com/google.cloud.audit.AuditLog",
      "authenticationInfo": {
        "principalEmail": "test@google.com"
      },
      "authorizationInfo": [
        {
          "granted": true,
          "permission": "pubsub.topics.create",
          "resource": "projects/test",
          "resourceAttributes": {}
        }
      ],
      "methodName": "google.pubsub.v1.Publisher.CreateTopic",
      "request": {
        "@type": "type.googleapis.com/google.pubsub.v1.Topic",
        "kmsKeyName": "",
        "name": "projects/test/topics/test-auditlogs-source"
      },
      "requestMetadata": {
        "callerIp": "2620:15c:10f:203:c404:c29d:a94d:dc1e",
        "callerSuppliedUserAgent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.88 Safari/537.36,gzip(gfe)",
        "destinationAttributes": {},
        "requestAttributes": {
          "auth": {},
          "time": "2020-01-07T20:56:30.520970037Z"
        }
      },
      "resourceLocation": {
        "currentLocations": [
          "us-west2"
        ]
      },
      "resourceName": "projects/test/topics/test-auditlogs-source",
      "response": {
        "@type": "type.googleapis.com/google.pubsub.v1.Topic",
        "messageStoragePolicy": {
            "us-west2"
          ]
        },
        "name": "projects/test/topics/test-auditlogs-source"
      },
      "serviceName": "pubsub.googleapis.com"
    },
    "receiveTimestamp": "2020-01-07T20:56:59.919268372Z",
    "resource": {
      "labels": {
        "project_id": "test",
        "topic_id": "projects/test/topics/test-auditlogs-source"
      },
      "type": "pubsub_topic"
    },
    "severity": "NOTICE",
    "timestamp": "2020-01-07T20:56:30.516179081Z"
  }
```

## What's Next

1. For integrating with Cloud Pub/Sub, see the [PubSub example](../pubsub/README.md).
1. For integrating with Cloud Storage see the [Storage example](../storage/README.md).
1. For more information about CloudEvents, see the [HTTP transport bindings documentation](https://github.com/cloudevents/spec).

## Cleaning Up

1. Delete the Scheduler

    ```shell
    kubectl delete -f ./auditlogssource.yaml
    ```
1. Delete the Service    
    
    ```shell
    kubectl delete -f ./event-display.yaml
    ```
