ARG UBI_IMAGE=registry.access.redhat.com/ubi7/ubi-minimal:latest
FROM ${UBI_IMAGE}
ENV KUBECONFIG /root/.kube/config
COPY bin/eks-operator /usr/bin/
ENTRYPOINT ["eks-operator"]
