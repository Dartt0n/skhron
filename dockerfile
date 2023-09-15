FROM golang:latest 
RUN mkdir /app 
ADD . /app/ 
WORKDIR /app 
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /bin/skhron . 


FROM scratch
COPY --from=0 /bin/skhron /bin/skhron
CMD ["/bin/skhron"]