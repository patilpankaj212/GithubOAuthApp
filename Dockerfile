FROM golang:1.14.1

RUN apt-get update && apt-get -y install git

WORKDIR /go/src/

RUN mkdir GitHubOAuthApp

COPY . GitHubOAuthApp/

WORKDIR /go/src/GitHubOAuthApp

EXPOSE 8000

CMD ["go", "run", "main.go"]