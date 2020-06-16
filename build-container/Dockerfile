FROM alpine
ARG binary
COPY ./$binary /app/datapi
RUN chmod 555 /app/datapi
WORKDIR /app
EXPOSE 3000
CMD ["/app/datapi", "-api"]