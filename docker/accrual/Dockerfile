FROM debian:bookworm-slim

WORKDIR /app

COPY ./cmd/accrual/accrual_linux_amd64 ./

ENTRYPOINT [ "/app/accrual_linux_amd64" ]