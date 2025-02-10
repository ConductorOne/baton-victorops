FROM gcr.io/distroless/static-debian11:nonroot
ENTRYPOINT ["/baton-victorops"]
COPY baton-victorops /