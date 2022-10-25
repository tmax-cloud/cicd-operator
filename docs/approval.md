# `Approval`

This guide lets you know how to use `Approval` feature.

* [Creating an Approval step](#creating-an-approval-step)
* [Reusing Approvers list](#reusing-approvers-list)
* [Send mail before/after approval](#send-mail-beforeafter-approval)
* [Approving/Rejecting the approval](#approvingrejecting-the-approval)
  * [Option.1 Using `cicdctl`](#option-1-using-cicdctl)
  * [Option.2 Using `curl`](#option-2-using-curl)

## Creating an `Approval` step
Add following 'approval' job before the job which needs an approval in `IntegrationConfig`
```yaml
- name: approval
  approval:
    approvers:
      - name: admin@tmax.co.kr
        email: sunghyun_kim3@tmax.co.kr
      - name: test@tmax.co.kr
      - name: system:serviceaccount:default:approver-account # Service account is also supported
- name: need-approval
  image: busybox
  after:
    - approval
```

## Reusing Approvers list
1. Create approvers list ConfigMap
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: approver-test
data:
  # <User name>=<Email address> form (email is optional)
  approvers: |
    admin@tmax.co.kr=sunghyun_kim3@tmax.co.kr
    test@tmax.co.kr
```

2. Add following 'approval' job before the job which needs an approval in `IntegrationConfig`
```yaml
- name: approval
  approval:
    approversConfigMap:
      name: approver-test
    approvers: # You can use both approvers & approversConfigMap
      - name: system:serviceaccount:default:approver-account
- name: need-approval
  image: busybox
  after:
     - approval
```

## Send mail before/after approval
Enable email feature, following the [installation guide](./installation.md#enabling-email-feature)
```yaml
- name: approval
  approval:
    approvers:
      - name: admin@tmax.co.kr
        email: sunghyun_kim3@tmax.co.kr
      - name: test@tmax.co.kr
      - name: system:serviceaccount:default:approver-account
    requestMessage: Please approve this! # Message to be sent via email when the Approval is created
- name: need-approval
  image: busybox
  after:
     - approval
```

## Approving/Rejecting the `Approval`
1. Find the requested user's token.  
   If you are using ServiceAccount for the user, you can find your token with following command
   ```bash
   SERVICE_ACCOUNT=<Name of the service account>
   kubectl get secret $(kubectl get serviceaccount $SERVICE_ACCOUNT -o jsonpath='{.secrets[].name}') -o jsonpath='{.data.token}' | base64 -d
   ```

2. Run API call to Kubernetes API server
   ### Option. 1 Using `cicdctl`
   ```bash
   TOKEN=<Token got from 1.>
   
   APPROVAL=<Name of the Approval object>
   NAMESPACE=<Namespace where the Approval exists>
   
   cicdctl --token $TOKEN -n $NAMESPACE [approve|reject] $APPROVAL <Reason of the decision>
   ```

   ### Option. 2 Using `curl`
   ```bash
   KUBERNETES_API_SERVER=<Kubernetes api server host:port>
   TOKEN=<Token got from 1.>
   
   APPROVAL=<Name of the Approval object>
   NAMESPACE=<Namespace where the Approval exists>
   
   DECISION=[approve|reject]
   REASON=<Reason of the decision>
   
   curl -k -X PUT \
   -H "Authorization: Bearer $TOKEN" \
   -d "{\"reason\": \"$REASON\"}"
   "$KUBERNETES_API_SERVER/apis/cicdapi.tmax.io/v1/namespaces/$NAMESPACE/approvals/$APPROVAL/$DECISION"
   ```
