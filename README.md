# ldap-jwt-go
Lightweight Go(golang) based Web service that provides user authentication against LDAP server (I tested only against [OpenLDAP](https://www.openldap.org/) server, but I think you can also use Active Directory) credentials and returns a JSON Web Token.

I was inspired by lucianweber's [ldap-jwt](https://github.com/lucianweber/ldap-jwt) to create this.

# How to use

## 1. Build runtime image

```shell
docker build -t michibiki.io/ldap-jwt-go:latest ./
```

### Optinal: Build development image
- This process is not required
- If you want to develop or modify this web service, I reccomend to use development image

```shell
docker build --target=development -t michibiki.io/ldap-jwt-go-dev:latest ./
```

## docker image size
|type|size|base image|
|:--|--:|:--|
|development|1.31GB|[golang:latest](https://hub.docker.com/_/golang)|
|runtime|116MB|[debian:bullseye-slim](https://hub.docker.com/_/debian)|

## 2. Set Environment Variables for LDAP authentication

- Edit a docker-compose.yml or use [Portainer](https://www.portainer.io)
- This Web service use "redis" server

|name|required|default|detail|
|:--|:-:|:--|:--|
|LDAP_PROTOCOL||LDAP|LDAP, LDAPS, START_TLS are supported|
|LDAP_HOST||localhost|LDAP server hostname or ipaddress|
|LDAP_PORT||389|LDAP server listen port|
|LDAP_SKIPVERIFY||false|Whether to check for SSL certificate is valid or not|
|LDAP_BIND_DN|v||Bind UserDN|
|LDAP_BIND_PASSWORD|v||Bind Password|
|LDAP_BASE_DN|v||search base for user and group|
|LDAP_FILTER_USER||(&(objectClass=posixAccount)(uid=%s))|filter for search userid|
|LDAP_FILTER_GROUP||(&(objectClass=groupOfNames)(member=%s))|filter for search user groups|
|REDIS_HOST||redis:6379|Redis server hostname and port|
|ACCESS_TOKEN_EXPIRE||15|Valid period of the access token (minites)|
|REFRESH_TOKEN_EXIPIRE||10080|Valid period of the refresh token (minites)|

## 3. Start Web service

```shell
docker-compose up
```

# Endpoints

## /v1/authorize

### Payload

```json
{
    "username": "exampleuser",
    "password": "password"
}
```

### Responce

- "expire_in" means how long the "access_token" is valid (seconds.)

```json
{
    "access_token": "eyJhbGciOiJSUzUxMiIsIn...",
    "expire_in": 899,
    "refresh_token": "eyJhbGciOiJSUzUxMiIsIn...",
    "token_type": "Bearer"
}
```

## /v1/verify

### Payload

```json
{
    "access_token": "eyJhbGciOiJSUzUxMiIsIn...",
}
```

### Responce

- "expire_in" means how long the "access_token" is valid (seconds.)

```json
{
    "expire_in":803,
    "user":{
        "DN":"cn=Taro Yamada,ou=users,dc=example,dc=com",
        "Id":"taro",
        "Groups":[
            "cn=users,ou=groups,dc=example,dc=com",
            "cn=guests,ou=groups,dc=example,dc=com"
        ]
    }
}
```

## /v1/refresh

- You are able to refresh your "access_token" when it expires.
- If you refresh the "access_token", the "refresh_token" will also be updated.
- The "refresh_token" can only be used once (One-time token).

### Payload

```json
{
    "refresh_token": "eyJhbGciOiJSUzUxMiIsIn...",
}
```

### Responce

- "expire_in" means how long the "refresh_token" is valid (seconds.)

```json
{
    "access_token":"eyJhbGciOiJSUzUxMiIsIn...",
    "expire_in":604799,
    "refresh_token":"eyJhbGciOiJSUzUxMiIsIn...",
    "token_type":"Bearer"
}
```

## /v1/deauthorize

- Disable the "access_token" and the "refresh_token" that associated with "access_token"
### Payload

```json
{
    "access_token": "eyJhbGciOiJSUzUxMiIsIn..."
}
```

### Responce

```json
{
    "result":true
}
```

# TODO
- Write a test code

# License
The MIT License (MIT)

Copyright (c) 2021 [Michibiki Technical Advisory Corporation](https://michibiki.io)

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
