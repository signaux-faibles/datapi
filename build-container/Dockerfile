FROM ubuntu:latest
RUN apt-get update && apt-get install -y python3 python3-dev python3-pip libc6-dev libxml2 libxml2-dev libxslt-dev gcc
RUN pip3 install docx-mailmerge==0.5.0
ARG datapiDir
COPY ./docxify3.py /app/docxify3.py
COPY ./template.docx /app/template.docx
COPY ./$datapiDir/datapi /app/datapi
COPY ./$datapiDir/migrations/ /app/migrations
RUN chmod 555 /app/datapi
WORKDIR /app
EXPOSE 3000
CMD ["/app/datapi"]
