# Lambda EKS Test

Test for accessing an EKS cluster from a Lambda function.

## Dependencies

### `aws-iam-authenticator`

Authentication against an EKS cluster requires the [`aws-iam-authenticator`](https://github.com/kubernetes-sigs/aws-iam-authenticator) executable in the Lambda execution environment.

- Build `aws-iam-authenticator` locally:

    ~~~bash
    GOOS=linux go build github.com/kubernetes-sigs/aws-iam-authenticator/cmd/aws-iam-authenticator
    ~~~

- Include the created `aws-iam-authenticator` executable in ZIP file for Lambda function

### *kubeconfig* File

Authentication against EKS cluster requires a *kubeconfig* file in the Lambda execution environment.

- Create *kubeconfig* file for target cluster locally:

    ~~~bash
    aws eks update-kubeconfig --name <ClusterName> --kubeconfig kubeconfig
    ~~~

- In the created file `kubeconfig`, change path to `aws-iam-authenticator` (in the YAML field `users.user.exec.command`) to a relative path according to file location from above
- Include `kubeconfig` file in ZIP file for Lambda function

## Lambda Function Code

Uses the Kubernetes [Go client library](https://github.com/kubernetes/client-go) to access the EKS cluster (API server).

Through the Go client library, read the `kubeconfig` file, and then make request to the EKS cluster.

A *kubeconfig* file for an EKS cluster requires to use the [`users.user.exec`](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#client-go-credential-plugins) feature for client authentication. Consequently, the used Kubernetes client library must understand this feature.

Currently, most Kubernetes client libraries don't yet have support for this feature. The Go client library has, which is the reason that we use the Go client library, which is the reason that we use a Go Lambda function.
