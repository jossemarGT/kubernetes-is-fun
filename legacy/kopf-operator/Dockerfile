FROM python:3.11

WORKDIR /src

COPY requirements.txt /src/
RUN pip install -r requirements.txt

COPY . /src/

CMD kopf run --all-namespaces /src/legacy-operator.py
