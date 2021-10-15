FROM python:3.10
#FROM alpine:3.12.0
#RUN apk add --no-cache python3 python3-dev musl-dev libxml2 libxml2-dev libxslt libxslt-dev gcc
#RUN python3 -m ensurepip
#RUN pip3 install docx-mailmerge==0.5.0
ARG datapiDir
COPY ./docxify3.py /app/docxify3.py
COPY ./template.docx /app/template.docx
COPY ./$datapiDir/datapi /app/datapi
COPY ./$datapiDir/migrations/ /app/migrations
RUN chmod 555 /app/datapi
WORKDIR /app
EXPOSE 3000
CMD ["/app/datapi"]
