FROM alpine:3.19

RUN apk add --no-cache curl unzip

RUN curl -sSL -o /tmp/tigerbeetle.zip https://linux.tigerbeetle.com \
    && unzip -o /tmp/tigerbeetle.zip -d /tmp \
    && mv /tmp/tigerbeetle /usr/local/bin/ \
    && chmod +x /usr/local/bin/tigerbeetle \
    && rm /tmp/tigerbeetle.zip

RUN mkdir -p /data
VOLUME /data

COPY start.sh /start.sh
RUN sed -i 's/\r$//' /start.sh && chmod +x /start.sh

EXPOSE 3000
CMD ["/start.sh"]
