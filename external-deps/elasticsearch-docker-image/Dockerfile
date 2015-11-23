FROM elasticsearch:2.0.0
MAINTAINER "André Martins <aanm90@gmail.com>"

ADD external-deps /tmp/external-deps

RUN /usr/share/elasticsearch/bin/plugin install \
    file:///tmp/external-deps/license-2.0.0.zip && \
    /usr/share/elasticsearch/bin/plugin install \
    file:///tmp/external-deps/watcher-2.0.0.zip && \
    /usr/share/elasticsearch/bin/plugin install \
    file:///tmp/external-deps/marvel-agent-2.0.0.zip

RUN cp /tmp/external-deps/elasticsearch.yml /usr/share/elasticsearch/config/elasticsearch.yml

RUN chown elasticsearch.elasticsearch /usr/share/elasticsearch/config/elasticsearch.yml
