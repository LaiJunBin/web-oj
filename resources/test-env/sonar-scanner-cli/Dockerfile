FROM sonarsource/sonar-scanner-cli:11.3

WORKDIR /project
USER root

COPY --chown=scanner-cli:scanner-cli ./run-test.sh /project/run-test.sh
RUN chmod +x /project/run-test.sh && chown -R scanner-cli:scanner-cli /project

CMD []
ENTRYPOINT ["tail", "-f", "/dev/null"]
