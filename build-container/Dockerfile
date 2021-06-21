FROM alpine:3.12.0
RUN apk add --no-cache python2 python2-dev musl-dev libxml2 libxml2-dev libxslt libxslt-dev gcc
RUN python -m ensurepip
RUN pip install docx-mailmerge
ARG datapiDir
COPY ./docxify.py /app/docxify.py
COPY ./template.docx /app/template.docx
COPY ./$datapiDir/datapi /app/datapi
COPY ./$datapiDir/migrations/ /app/migrations
RUN chmod 555 /app/datapi
WORKDIR /app
EXPOSE 3000
CMD ["/app/datapi"]
