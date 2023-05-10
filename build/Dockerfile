FROM alpine:3.17.3
ARG VERSION
ARG TARGETPLATFORM
ENV SOURCE_DATA_PATH=/data
LABEL io.soutwinds.uri="www.southwinds.io"
LABEL io.soutwinds.author="Southwinds Tech Ltd"
LABEL io.southwinds.maintainer="admin@southwinds.io"
LABEL io.southwinds.product="Source Configuration Service"
LABEL io.southwinds.version=$VERSION
LABEL io.southwinds.LICENSE="Enterprise Lisence Agreement"
RUN addgroup --system play && \
    adduser --system play --ingroup play && \
    apk update && \
    apk add ca-certificates && \
    update-ca-certificates && \
    rm -rf /var/cache/apk/*
WORKDIR /app
COPY $TARGETPLATFORM/source /app
USER play
VOLUME $SOURCE_DATA_PATH
EXPOSE 8080
CMD ["/app/source"]