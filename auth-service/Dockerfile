FROM scratch

COPY ca-bundle.crt /etc/ssl/certs/ca-certificates.crt
COPY auth-service /
COPY register-account /

ENTRYPOINT ["/auth-service"]
CMD [""]
