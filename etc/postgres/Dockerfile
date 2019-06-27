FROM postgres:11.3

ENV DEBIAN_FRONTEND noninteractive
RUN set -x \
    && apt-get update \
    && apt-get install --no-install-recommends -y \
        postgresql-11-postgis-2.5 \
        postgresql-11-postgis-2.5-scripts \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*
