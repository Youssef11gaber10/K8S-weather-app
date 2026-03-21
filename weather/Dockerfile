

FROM python:3.11 AS build

WORKDIR /app 

COPY requirements.txt .

RUN python -m venv /opt/venv && \
    . /opt/venv/bin/activate && \
    pip install --no-cache-dir -r requirements.txt && \
    /opt/venv/bin/pip install --no-cache-dir gunicorn 
    
FROM python:3.11-slim

COPY --from=build /opt/venv /opt/venv 
# the venv must be aligh with /app
WORKDIR /app
COPY . .
ENV PATH="$PATH:/opt/venv/bin" 
# ENV PATH="/opt/venv/bin:$PATH" # best practice 

EXPOSE 5000

# CMD [ "uvicorn", "app:app" ]
#we run with uvicorn cause this is asgi app 
# CMD [ "gunicorn","-K","uvicorn.workers.UvicornWorker", "app:app" ]
# CMD [ "gunicorn","-k", "uvicorn.workers.UvicornWorker","--bind", "0.0.0.0:6000","app:app"]
#we use the gunicorn here as process manager to run the app and manage the workers and we use uvicorn as worker to run the app because it's asgi app
#app is the name of the file and another app is the name of the variable that have the fastapi instance in it @app.route('/')
# no need for "--bind", "0.0.0.0:5000" cause he did it in code with uvicorn.run(app, host=") 
# the bind is important for gunicorn to bind the app to the port and make it accessible from outside the container but  with unicorn not important to write --host or --port its already in code
# CMD ["uvicorn", "main:app" ]
CMD ["gunicorn", "main:app", "--bind", "0.0.0.0:5000", "--workers", "3"]