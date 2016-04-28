FROM scratch

COPY ca-bundle.crt /etc/ssl/certs/ca-certificates.crt
COPY bookmark /
COPY public /public

ENTRYPOINT ["/bookmark"]
CMD [""]
