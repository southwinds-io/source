#
#    Source Configuration Service
#    © 2022 Southwinds Tech Ltd - www.southwinds.io
#
FROM registry.access.redhat.com/ubi8/ubi-minimal

ARG UNAME=source
ARG VERSION=0.1

LABEL io.soutwinds.uri="www.southwinds.io"
LABEL io.soutwinds.author="Southwinds Tech Ltd"
LABEL io.southwinds.maintainer="skipper@southwinds.io"
LABEL io.southwinds.product="Source Configuration Service"
LABEL io.southwinds.LICENSE="Apache 2"
LABEL io.southwinds.version=$VERSION

ENV UID=1000
ENV GID=1000
ENV SW_SRC_DATA_PATH=/data

RUN microdnf update --disablerepo=* --enablerepo=ubi-8-appstream --enablerepo=ubi-8-baseos -y && \
    microdnf install shadow-utils.x86_64 && \
    groupadd -g $GID -o $UNAME && \
    # -M create the user with no /home
    useradd -M -u $UID -g $GID $UNAME && \
    rm -rf /var/cache/yum && \
    mkdir ${SW_SRC_DATA_PATH} && chown -R ${UNAME} ${SW_SRC_DATA_PATH} && \
    microdnf clean all

WORKDIR /app

COPY src ./

USER $UNAME

VOLUME $SW_SRC_DATA_PATH

ENTRYPOINT ["sh", "-c", "/app/src"]

EXPOSE 8080/tcp
