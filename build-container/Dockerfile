FROM alpine
ARG datapiDir
COPY ./$datapiDir/datapi /app/datapi
COPY ./$datapiDir/migrations/ /app/migrations
RUN chmod 555 /app/datapi
WORKDIR /app
EXPOSE 3000
CMD ["/app/datapi"]