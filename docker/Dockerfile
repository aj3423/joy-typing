FROM alphacep/kaldi-vosk-server:latest

ENV MODEL_VERSION 0.22
RUN mkdir /opt/vosk-model-en \
   && cd /opt/vosk-model-en \
   && wget -q http://alphacephei.com/kaldi/models/vosk-model-en-us-${MODEL_VERSION}-lgraph.zip \
   && unzip vosk-model-en-us-${MODEL_VERSION}-lgraph.zip \
   && mv vosk-model-en-us-${MODEL_VERSION}-lgraph model \
   && rm -rf vosk-model-en-us-${MODEL_VERSION}-lgraph.zip

ENV VOSK_SERVER_PORT 2701
EXPOSE 2701
WORKDIR /opt/vosk-server/websocket
CMD [ "python3", "./asr_server.py", "/opt/vosk-model-en/model" ]
