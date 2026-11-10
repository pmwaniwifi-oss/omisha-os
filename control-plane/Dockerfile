FROM python:3.12-slim
WORKDIR /app
COPY app.py .
RUN useradd -u 10001 -m app && mkdir /data && chown app:app /data
USER app
EXPOSE 8080
CMD ["python","app.py"]
