# stage 1: build
FROM golang:1.23.3 as build
LABEL stage=intermediate
WORKDIR /app
COPY . .
RUN make build

# stage 2: scratch
FROM alpine:3.20.3 as scratch
COPY --from=build /app/bin/katok /bin/katok
USER        nobody
ENTRYPOINT  ["/bin/katok"]
CMD [ "" ]