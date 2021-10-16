#####################################
########## For Development ##########
#####################################
FROM golang:latest as development

RUN apt-get update \
 && apt-get install -y --no-install-recommends \
    gosu \
 && apt-get -y clean \
 && rm -rf /var/lib/apt/lists/*

RUN mkdir -p /opt/go && \
    mkdir -p /opt/go/private && \
    groupadd go && \
    useradd -m -s /bin/bash -g go go && \
    chown go:go /opt/go -R

COPY entrypoint.sh /opt
RUN chmod go+x /opt/entrypoint.sh

# Envs
ENV LDAP_PROTOCOL=LDAP \
    LDAP_HOST=localhost \
    LDAP_PORT=389 \
    LDAP_SKIPVERIFY=false \
    LDAP_BIND_DN=cn=readonly,dc=example,dc=com \
    LDAP_BIND_PASSWORD=readonly \
    LDAP_BASE_DN=dc=example,dc=com \
    LDAP_FILTER_USER=(&(objectClass=posixAccount)(uid=%s)) \
    LDAP_FILTER_GROUP=(&(objectClass=groupOfNames)(member=%s)) \
    REDIS_HOST=localhost:6379 \
    ACCESS_TOKEN_EXPIRE=15 \
    REFRESH_TOKEN_EXIPIRE=10080

WORKDIR /opt/go

COPY . .

# require tools
RUN GOBIN=/tmp/ go get github.com/go-delve/delve/cmd/dlv@master && \
    mv /tmp/dlv $GOPATH/bin/dlv-dap && \
    go install golang.org/x/tools/gopls@latest && \
    go mod tidy && go build -o ./ldap-jwt.go ./main.go

VOLUME ["/opt/go"]

EXPOSE 80

ENTRYPOINT ["/opt/entrypoint.sh"]

CMD ["/bin/bash"]

#################################
########## For RUNTIME ##########
#################################
FROM debian:bullseye-slim as runtime

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    openssh-client gosu && \
    apt-get autoremove -y && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

RUN mkdir -p /opt/go && \
    mkdir -p /opt/go/private

COPY --from=development /opt/go/ldap-jwt.go /opt/go/ldap-jwt.go

RUN groupadd go && \
    useradd -m -s /bin/bash -g go go && \
    chown go:go /opt/go -R

COPY entrypoint.sh /opt
RUN chmod go+x /opt/entrypoint.sh

# Envs
ENV GIN_MODE=release \
    LDAP_PROTOCOL=LDAP \
    LDAP_HOST=localhost \
    LDAP_PORT=389 \
    LDAP_SKIPVERIFY=false \
    LDAP_BIND_DN=cn=readonly,dc=example,dc=com \
    LDAP_BIND_PASSWORD=readonly \
    LDAP_BASE_DN=dc=example,dc=com \
    LDAP_FILTER_USER=(&(objectClass=posixAccount)(uid=%s)) \
    LDAP_FILTER_GROUP=(&(objectClass=groupOfNames)(member=%s)) \
    REDIS_HOST=localhost:6379 \
    ACCESS_TOKEN_EXPIRE=15 \
    REFRESH_TOKEN_EXIPIRE=10080

WORKDIR /opt/go

USER go

EXPOSE 80

VOLUME ["/opt/go/private"]

ENTRYPOINT ["/opt/entrypoint.sh"]

CMD ["/opt/go/ldap-jwt.go"]
