FROM ubuntu:24.04
RUN apt-get update && \
  apt-get -y install curl git jq && \
  apt-get clean && \
  rm -rf /var/lib/apt/lists /var/cache/apt/archives && \
  curl -LsSfO "https://dl.k8s.io/release/$(curl -LsSf https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" && \
  chmod +x ./kubectl && \
  mv kubectl /usr/local/bin

WORKDIR /root

COPY generate-github-installation-access-token.sh wait-for-resource.sh git-push.sh deploy-step-argocd.sh .

CMD ["./deploy-step-argocd.sh"]
