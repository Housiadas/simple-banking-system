ARG UID
ARG GID

FROM migrate/migrate

WORKDIR /app

COPY .. .

USER $UID:$GID

ENTRYPOINT ["migrate"]
