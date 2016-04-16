FROM alpine:3.3

COPY ca-bundle.crt /etc/ssl/certs/ca-certificates.crt
COPY feedreaderd /
COPY sources.txt /
COPY */templates/*.html /templates/

ENV TEMPLATES=/templates/*.html

ENTRYPOINT ["/feedreaderd"]
CMD [""]
