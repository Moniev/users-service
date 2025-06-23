FROM python:3.9-slim-buster

WORKDIR /app

COPY requirements.txt .

RUN pip install --no-cache-dir -r requirements.txt

COPY app/ ./app/

EXPOSE 8000

CMD ["gunicorn", "--bind", "0.0.0.0:8000", "-w", "4", "-k", "uvicorn.workers.UvicornWorker", "app.main:app"]