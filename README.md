# Lambda EKS Test

Accessing an EKS cluster from a Lambda function.

## What is it?

This is a Go Lambda function that accesses an EKS cluster. In particular, the Lambda function creates a new [deployment](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.12/#deployment-v1-apps) in an existing EKS cluster.

The program uses the [Kubernetes Go client library](https://github.com/kubernetes/client-go) (client-go) to access the Kubernetes cluster (more precisely, to talk to the API server of the cluster). Read [below](why-go) why we use Go.

A Lambda function that accesses an EKS cluster requires some additional steps which are summarised in the [Requirements](requirements) section. You have to work through this section before deploying the Lambda function.

The code in this repository requires an already existing EKS cluster.


## Compilation

Since we use Go, the Lambda handler function has to be compiled before deployment (what's deployed is a statically linked binary). Furthermore, the code has to be compiled specifically for the [Lambda execution environment](https://docs.aws.amazon.com/lambda/latest/dg/current-supported-versions.html) platform (which is Linux). Compiling for any target platform can be achieved with Go [cross compiling](https://dave.cheney.net/2015/08/22/cross-compilation-with-go-1-5).

To cross-compile the code for the Lambda execution environment, use the following command (already defined in [`build-handler.sh`](build-handler.sh)):

~~~bash
GOOS=linux go build handler.go
~~~

## Deployment

The Lambda application is defined with [SAM](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/what-is-sam.html) (in [`template.yml`](template.yml)).

Deploy the Lambda application to AWS with the following SAM CLI commands (already defined in [`deploy.sh`](deploy.sh])):

~~~bash
sam package --template-file template.yml --output-template-file package.yml --s3-<SOME_BUCKET>
sam deploy --template-file package.yml --capabilities CAPABILITY_IAM --stack-name lambda-eks-test
~~~

## Local Testing

SAM allows to execute Lambda functions locally (they will run in a Docker container that simulates the Lambda execution environment).

To run the function locally, use the following SAM CLI command (already defined in [`local.sh`](local.sh)):

~~~bash
sam local invoke --no-event LambdaEksTestFunction
~~~


## Requirements

### `aws-iam-authenticator`

Authentication against an EKS cluster requires the `[aws-iam-authenticator](https://github.com/kubernetes-sigs/aws-iam-authenticator)` executable (this executable is referenced in the *kubeconfig* file, see next section). That means, `aws-iam-authenticator` must be present in the Lambda execution environment.

1. Install `aws-iam-authenticator` on your machine, if you haven't already:

    ~~~bash
    go get -u github.com/kubernetes-sigs/aws-iam-authenticator/cmd/aws-iam-authenticator
    ~~~

2. Build `aws-iam-authenticator` for the [Lambda execution enviroment](https://docs.aws.amazon.com/lambda/latest/dg/current-supported-versions.html):

    ~~~bash
    GOOS=linux go build github.com/kubernetes-sigs/aws-iam-authenticator/cmd/aws-iam-authenticator
    ~~~

The created `aws-iam-authenticator` must be included in the ZIP file for the Lambda function.

### *kubeconfig* File

To make requests to a Kubernetes cluster, a [*kubeconfig*](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/) file is required.  The *kubeconfig* file must be present in the execution environment of the Lambda function. The *kubeconfig* file is read by the Go client library.

1. Create a *kubeconfig* file for the target cluster:

    ~~~bash
    aws eks update-kubeconfig --name <ClusterName> --kubeconfig kubeconfig
    ~~~

2. In the created file, replace `aws-iam-authenticator` with `./aws-iam-authenticator` (in the `users.user.exec.command` field)
    - This is necessary to make the command executable in the Lambda execution enviroment, because in the Lambda execution environment `aws-iam-authenticator` is not in the `PATH`

The created `kubeconfig` file must be included in the ZIP file for the Lambda function.

### Authentication for Lambda execution role

The Go client library uses the IAM role assigned to the Lambda function to authenticate to the EKS cluster (the determination of the role and encoding in the bearer authentication token is done by the `aws-iam-authenticator`).

To make the cluster recognise and authenticate requests coming from the Lambda function, we must add this role to the `aws-auth` ConfigMap of the cluster.

1. Open the `aws-auth` ConfigMap for editing:

    ~~~bash
    kubectl edit -n kube-system configmap/aws-auth
    ~~~

2. Append the following data to the already existing value of the `data.mapRoles` field in the ConfigMap (replace `<LambdaRoleARN>` with the ARN of the execution role of the Lambda function): 

    ~~~yaml
    mapRoles: |
      - rolearn: <LambdaRoleARN>
        username: lambda
    ~~~

3. Save the file (the changes are automatically applied to the cluster)    


### Authorisation for Lambda execution role

At this point, requests from the Lambda function to the cluster get *authenticated*, but the specific Kubernetes action requested by our program (create deployments) does not yet get *authorised*.

In the authentication step, we map requests from the Lambda function to an internal Kubernetes user called `lambda`. This is a custom user that we just invented and it does not have any permissions by default.

We have to assign permission to create deployments to the `lambda` user.

EKS uses the native Kubernetes RBAC authorisation system. In RBAC, permissions are defined as [Role](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.12/#role-v1-rbac-authorization-k8s-io) objects and assigned to users with [RoleBinding](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.12/#rolebinding-v1-rbac-authorization-k8s-io) objects. Role and RoleBinding are ordinary Kubernetes API resource objects, and you can define them like other API resource objects.

1. Define the following Role and RoleBinding (already exists in file [permissions.yml](permissions.yml)):

    ~~~yaml
    kind: Role
    apiVersion: rbac.authorization.k8s.io/v1
    metadata:
      name: lambda-create-deployments
      namespace: default
    rules:
    - apiGroups: ["apps"]
      resources: ["deployments"]
      verbs: ["create"]
    ---
    kind: RoleBinding
    apiVersion: rbac.authorization.k8s.io/v1
    metadata:
      name: lambda-create-deployments
      namespace: default
    subjects:
      - kind: User
        name: lambda
        apiGroup: rbac.authorization.k8s.io
    roleRef:
      kind: Role
      name: lambda-create-deployments
      apiGroup: rbac.authorization.k8s.io
    ~~~

2. Create the Role and RoleBinding:

    ~~~bash
    kubectl apply -f permissions.yml
    ~~~

Now, requests from the Lambda function to create a deployment will be authorised. But any other request, for example, to list deployments, will be denied by the authorisation system. That's exactly what we want to guarantee the *principle of least privilege*.

If you extend the Lambda function to do other Kubernetes actions, you have to adapt the RBAC permissions through Role and RoleBinding objects accordingly.

## Why Go?

The client-side part of the EKS authentication mechanism makes use of a Kubernetes feature called *exec provider* or *credentials plugin*, which is defined [here](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#client-go-credential-plugins). The feature defines an `exec` section in the *kubeconfig* file. This section defines an external command that returns the credentials to authenticate to the cluster (in the case of EKS this external command is `aws-iam-authenticator` and the returned credential is a bearer token that encodes an IAM identity).

Support for this feature must be implemented in the different [Kubernetes client libraries](https://kubernetes.io/docs/reference/using-api/client-libraries/) (because the client libraries read the *kubeconfig* file). However, at the moment, most client libraries to not yet support this feature. The Go client library does, and this is the reason that we use Go and a Go Lambda function.

