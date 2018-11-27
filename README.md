# Lambda EKS Test

Test for accessing an EKS cluster from a Lambda function.

## Dependencies

### `aws-iam-authenticator`

Authentication against EKS cluster requires [`aws-iam-authenticator`](https://github.com/kubernetes-sigs/aws-iam-authenticator).

- Build `aws-iam-authenticator` locally:

    ~~~bash
    GOOS=linux go build github.com/kubernetes-sigs/aws-iam-authenticator/cmd/aws-iam-authenticator
    ~~~

- Include created binary `aws-iam-authenticator` in ZIP file for Lambda function

### *kubeconfig* File

Authentication against EKS cluster requires a *kubeconfig* file.

- Create *kubeconfig* file for target cluster locally:

    ~~~bash
    aws eks update-kubeconfig --name <ClusterName> --kubeconfig kubeconfig
    ~~~

- In created file `kubeconfig`, adapt path to `aws-iam-authenticator` in the YAML field `users.user.exec.command`
- Include `kubeconfig` file in ZIP file for Lambda function

## Lambda Function Code

Uses the Kubernetes [Go client library](https://github.com/kubernetes/client-go) to access the EKS cluster (API server). For this reason, it is a Go Lambda function.

The Go client library reads the `kubeconfig` file and then allows to make request to the EKS cluster API server.

Reading a *kubeconfig* file for EKS requires the Kubernetes client library to understand the [`users.user.exec`](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#client-go-credential-plugins) syntax in the *kubeconfig* file. Currently, most Kubernetes client libraries don't yet have support for this feature. The Go client library has, which is the reason that we use the Go client library, which is the reason that we use a Go Lambda function.
