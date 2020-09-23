# Builder Image
# ---------------------------------------------------
FROM dimaskiddo/alpine:go-1.12 AS go-builder

WORKDIR /usr/src/app

COPY . ./

RUN go mod download \
    && CGO_ENABLED=0 GOOS=linux go build -a -o dist/go-whatsapp *.go


# Final Image
# ---------------------------------------------------
FROM dimaskiddo/alpine:base
LABEL MAINTAINER Tarikh Agustia Ijudin <agustia.tarikh150@gmail.com>

ARG SERVICE_NAME="go-whatsapp-rest"

ENV PATH $PATH:/opt/${SERVICE_NAME}

WORKDIR /opt/${SERVICE_NAME}

RUN mkdir storage

COPY --from=go-builder /usr/src/app/dist/go-whatsapp ./go-whatsapp

CMD ["go-whatsapp"]