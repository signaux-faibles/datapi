FROM alpine
ARG datapiDir
COPY ./$datapiDir/datapi /app/
COPY ./$datapiDir/migrations/ /app/
RUN chmod 555 /app/datapi
WORKDIR /app
EXPOSE 3000
CMD ["/app/datapi"]