FROM alpine:3.18.2
ARG VERSION
ARG TARGETPLATFORM
ARG UNAME=source
ENV SOURCE_HOME=/home/$UNAME
ENV SOURCE_DATA_PATH=/data
ENV UID=1000
ENV GID=1000
LABEL io.soutwinds.uri="www.southwinds.io"
LABEL io.soutwinds.author="Southwinds Tech Ltd"
LABEL io.southwinds.maintainer="admin@southwinds.io"
LABEL io.southwinds.product="Lightweight Configuration Service"
LABEL io.southwinds.version=$VERSION
LABEL io.southwinds.LICENSE="Enterprise Lisence Agreement"
RUN addgroup --system $UNAME --gid $GID && \
    adduser --system $UNAME --uid $UID --ingroup $UNAME && \
    apk update && \
    apk add ca-certificates && \
    update-ca-certificates && \
    rm -rf /var/cache/apk/* && \
    mkdir -p $SOURCE_DATA_PATH && \
    chown -R $UNAME:$UNAME $SOURCE_DATA_PATH
WORKDIR $SOURCE_HOME
COPY $TARGETPLATFORM/source .
RUN chown -R $UNAME:$UNAME $SOURCE_HOME/source
USER $UNAME
VOLUME $SOURCE_DATA_PATH
EXPOSE 8080
CMD ["./source"]