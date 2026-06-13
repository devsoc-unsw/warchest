FROM ghcr.io/tigerbeetle/tigerbeetle:latest

# Clear the inherited ENTRYPOINT
ENTRYPOINT []

RUN mkdir -p /data
VOLUME /data

COPY start.sh /start.sh
RUN sed -i 's/\r$//' /start.sh && chmod +x /start.sh

EXPOSE 3000
CMD ["/start.sh"]
