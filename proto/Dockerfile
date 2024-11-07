FROM ubuntu
WORKDIR /home/proto
RUN apt-get update
RUN apt install -y protobuf-compiler
RUN apt install -y protoc-gen-go
ENTRYPOINT ["bash","-c","/home/proto/gen-proto.sh"]
#CMD ["bash"]