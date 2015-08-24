FROM python:2.7
MAINTAINER Ahmet Alp Balkan

RUN mkdir /app
WORKDIR /app
COPY requirements.txt /app/
RUN pip install -r requirements.txt

# Add simplegauges dependency
RUN git clone https://github.com/ahmetalpbalkan/simplegauges.git

COPY . /app/
ENTRYPOINT ["./taskhost.py"]
