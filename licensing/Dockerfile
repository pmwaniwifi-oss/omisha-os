FROM python:3.12-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY server.py .
RUN useradd -u 10001 -m app && mkdir -p /data && chown app:app /data
USER app
EXPOSE 8090
CMD ["python","server.py"]
